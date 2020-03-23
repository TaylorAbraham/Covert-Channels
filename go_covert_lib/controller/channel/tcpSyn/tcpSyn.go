package tcpSyn

import (
	"../embedders"
	"bytes"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"math/rand"
	"net"
	"time"
)

const (
	Buffer   = 0
	Protocol = 1
)

type WriteWaitCancel struct{}

func (o *WriteWaitCancel) Error() string {
	return "Write Cancelled"
}

type Config struct {
	// The configuration of an ipv4_tcp_sequence covert channel
	// This structure recognizes three IP-port pairs, the friend, the origin, and an optional bounce.
	// The friend is the node you are sending messages to.
	// The origin is where messages are sent from when they reach you.
	// If there in bounce mode, the bounce address and port are used by you to bounce
	// messages to your friend (or for your friend to bounce messages to you)
	// thus avoiding direct communication between you ande your friend.
	FriendIP   [4]byte
	OriginIP   [4]byte
	BounceIP   [4]byte
	FriendPort uint16
	OriginPort uint16
	BouncePort uint16
	// In bounce mode, packets are not sent to the friend directly. Instead, they are sent
	// to a bouncer running a TCP socket on the origin IP-port. The packet SYN has the source IP-port
	// spoofed as the friend IP-port, so that when the bouncer replies with a SYN-ACK packet it will
	// be transmitted to your friend.
	Bounce bool
	// The delimiter to use to deliniate messages. Currently it is either no deliniation (Delim::None)
	// or delinieating by a TCP packet with a specific flag (Delim::Protocol).
	// Default is Delim::Protocol.
	Delimiter uint8
	Encoder   TcpEncoder
	// A function to retrieve a delay to implement between sent packets. By default this
	// function returns a delay of 0 ms, but users can set it to a longer time or even to
	// their favourite distribution.
	GetDelay func() time.Duration

	// Timeouts for reads and writes
	// These are the inter packet timeout, which will always translate to the inter byte Timeout
	// since at this time each packet may only contain one byte
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

// A TCP covert channel
type Channel struct {
	conf        Config
	rawConn     *ipv4.RawConn
	writeCancel chan bool
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	c := &Channel{conf: conf, writeCancel: make(chan bool)}
	if c.conf.Encoder == nil {
		c.conf.Encoder = &embedders.TcpIpSeqEncoder{}
	}

	conn, err := net.ListenPacket("ip4:6", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	c.rawConn, err = ipv4.NewRawConn(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return c, nil
}

// Receive a covert message
// data is a buffer for the message.
// If the channel is in protocol delimiter mode then this function
// will return as soon as it receives the proper delimiter.
// Otherwise, it will wait until the buffer is filled.
// We return the number of bytes received even if an error is encountered,
// in which case data will have valid received bytes up to that point.
func (c *Channel) Receive(data []byte) (uint64, error) {

	// We must expand out the input storage array to
	// the correct size to potentially handle variable size inputs
	dataBuf, err := embedders.GetBuf(c.conf.Encoder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	if len(dataBuf) == 0 && c.conf.Delimiter == Buffer {
		return 0, nil
	}

	var (
		buf          []byte = make([]byte, 1024)
		prev_val     uint32 = 0
		current_val  uint32 = 0
		first        bool   = true
		saddr        [4]byte
		sport, dport uint16
		// There is guaranteed to be at least one space for a byte in the
		// data buffer at this point
		pos uint64 = 0
		// The time since the last packet arrived
		// Timeouts can occur due to the raw socket itself timing out,
		// however this will typically not happen on a normal system
		// since the raw socket will read any incoming tcp packet.
		// This timer is used to timeout if packets are received, but they
		// are not the correct type
		prevPacketTime time.Time
		maskIndex int = 0
	)

	// Figure out the expected source and destination IP address
	// These change depending on whether or not we are in bounce mode
	if c.conf.Bounce {
		saddr, sport, dport = c.conf.BounceIP, c.conf.BouncePort, c.conf.OriginPort
	} else {
		saddr, sport, dport = c.conf.FriendIP, c.conf.FriendPort, c.conf.OriginPort
	}

	prevPacketTime = time.Now()

	var (
		h *ipv4.Header
		p []byte
	)
readloop:
	for {
		h, p, _, err = c.readConn(buf)
		if err != nil {
			break readloop
		}
		tcph := layers.TCP{}
		if err = tcph.DecodeFromBytes(p, gopacket.NilDecodeFeedback); err == nil {
			// We check for the expected source IP, source port, and destination port
			if bytes.Equal(h.Src.To4(), saddr[:]) {
				if tcph.SrcPort == layers.TCPPort(sport) && tcph.DstPort == layers.TCPPort(dport) {
					if c.conf.Delimiter == Protocol {
						if (!c.conf.Bounce && tcph.ACK && !tcph.RST) || (c.conf.Bounce && tcph.RST) {
							break readloop
						}
					}
					// If in bounce mode, we extract the original sequence number
					// from the acknowledgement number
					if c.conf.Bounce {
						tcph.Seq = tcph.Ack - 1
					}
					prev_val = tcph.Seq

					// Check the expected flags
					if (c.conf.Bounce && tcph.SYN && tcph.ACK) || (!c.conf.Bounce && tcph.SYN && !tcph.ACK) {
						if first || current_val != prev_val {
							prev_val = current_val
							first = false

							var newBytes []byte
							newBytes, err = c.conf.Encoder.GetByte(*h, tcph, maskIndex)
							if err != nil {
								break readloop
							}
							maskIndex = embedders.UpdateMaskIndex(c.conf.Encoder.GetMask(), maskIndex)

							for _, b := range newBytes {
								// Make sure we don't overflow the buffer
								// The buffer can only overflow if we are in protocol delimiter mode
								// Once the buffer fills in protocol delimiter mode we
								// wait for the end packet. If a packet arrives that does not
								// have the expected flags then we will reach this statement
								// and be unable to put the new byte into the buffer
								// In that case we return with an error indicating buffer overflow
								// (see below)
								if pos < uint64(len(dataBuf)) {
									dataBuf[pos] = b
								} else if pos == uint64(len(dataBuf)) && c.conf.Delimiter == Protocol {
									// If we are using the protocol delimiter, then we can fill the full
									// buffer and then wait for the end packet. It is only if more bytes
									// are received that we must notify of an error
									err = errors.New("End packet not received, buffer full")
									break readloop
								}
								pos += 1
								// We have filled the buffer without protocol delimiter and we should return immediately
								if pos == uint64(len(dataBuf)) && c.conf.Delimiter != Protocol {
									break readloop
								}
							}
							prevPacketTime = time.Now()
						}
					}
				}
			}
		}
		if c.conf.ReadTimeout > 0 && time.Now().Sub(prevPacketTime) > c.conf.ReadTimeout {
			err = errors.New("Covert Packet Timeout")
			break readloop
		}
	}

	return embedders.CopyData(c.conf.Encoder.GetMask(), pos, dataBuf, data, err)
}

// Send a covert message
// data is the entire message that will be sent.
// The GetDelay function can be used to set a large
// inter packet delay to help obscure the communication.
// We return the number of bytes sent even if an error is encountered
func (c *Channel) Send(data []byte) (uint64, error) {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var (
		prevSequence uint32 = r.Uint32()
		saddr, daddr [4]byte
		sport, dport uint16
		num          uint64 = 0
		h            ipv4.Header
		p            []byte
		cm           ipv4.ControlMessage
		wait         time.Duration
		tcph         layers.TCP
		// We make it clear that the error always starts as nil
		err error = nil
		maskIndex int = 0
	)

	data, err = embedders.EncodeFromMask(c.conf.Encoder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	// The source and destination depend on whether or not we are in bounce mode
	if c.conf.Bounce {
		saddr, daddr, sport, dport = c.conf.FriendIP, c.conf.BounceIP, c.conf.FriendPort, c.conf.BouncePort
	} else {
		saddr, daddr, sport, dport = c.conf.OriginIP, c.conf.FriendIP, c.conf.OriginPort, c.conf.FriendPort
	}
readloop:
	for len(data) > 0 {
		h = createIPHeader(saddr, daddr)
		cm = createCM(saddr, daddr)

		h, tcph, data, err = c.createTcpHead(h, layers.TCP{SYN: true}, data, prevSequence, maskIndex)
		if err != nil {
			break readloop
		}
		maskIndex = embedders.UpdateMaskIndex(c.conf.Encoder.GetMask(), maskIndex)
		prevSequence = tcph.Seq

		p, err = createTcpHeadBuf(tcph, saddr, daddr, sport, dport)
		if err != nil {
			break readloop
		}

		if err = c.writeConn(&h, p, &cm); err != nil {
			break readloop
		}

		// If the user did not supply a GetDelay function,
		// we wait for 0 duration
		// Otherwise we wait for the specified amount of time
		// (recalculated in each loop in case it must follow a distribution)
		if c.conf.GetDelay == nil {
			wait = 0
		} else {
			wait = c.conf.GetDelay()
		}

		num += 1
		// We wait for the time specified by the user's GetDelay function
		// or until the user signals to cancel
		select {
		case <-time.After(wait):
		case <-c.writeCancel:
			err = &WriteWaitCancel{}
			break readloop
		}
	}

	// Readjust size to represent number of bytes sent
	num, err = embedders.GetSentSize(c.conf.Encoder.GetMask(), num, err)
	if err != nil {
		return num, err
	}

	// If we are using Protocol delimiting, we must send a final
	// ACK packet to alert the receiver that the message is done sending
	if c.conf.Delimiter == Protocol {
		h = createIPHeader(saddr, daddr)
		cm = createCM(saddr, daddr)

		// We want to still encode data so that this final packet looks
		// like others. To do that, we need to provide enough bytes
		// to fit the mask at this maskIndex
		// We create a buffer with the appropriate size and fill it with
		// random numbers
		var fakeBuf []byte = make([]byte, len(c.conf.Encoder.GetMask()[maskIndex]))
		for i := range fakeBuf {
			fakeBuf[i] = byte(r.Uint32())
		}

		h, tcph, _, err = c.createTcpHead(h, layers.TCP{ACK: true}, fakeBuf, prevSequence, maskIndex)
		if err != nil {
			return num, err
		}
		prevSequence = tcph.Seq

		p, err = createTcpHeadBuf(tcph, saddr, daddr, sport, dport)
		if err != nil {
			return num, err
		}

		if err = c.writeConn(&h, p, &cm); err != nil {
			return num, err
		}
	}

	return num, nil
}

// Closes the covert channel.
// If the covert channel is closed while a read or write is occuring then
// the function function will return with an OpCancel error
func (c *Channel) Close() error {
	// The write operation allows the user to specify
	// a delay between packets
	// We can't just rely on the raw connection being
	// closed, we must also be able to cancel this delay,
	// which is done with the writeCancel method
	select {
	case <-c.writeCancel:
	default:
		close(c.writeCancel)
	}
	return c.rawConn.Close()
}

// Read from a raw connection whil setting a timeout if necessary
func (c *Channel) readConn(buf []byte) (*ipv4.Header, []byte, *ipv4.ControlMessage, error) {
	if c.conf.ReadTimeout > 0 {
		c.rawConn.SetReadDeadline(time.Now().Add(c.conf.ReadTimeout))
	} else {
		// A deadline of zero means never timeout
		// The initial Time struct is zero
		c.rawConn.SetReadDeadline(time.Time{})
	}
	return c.rawConn.ReadFrom(buf)
}

// Write to a raw connection while setting a timeout if necessary
func (c *Channel) writeConn(h *ipv4.Header, p []byte, cm *ipv4.ControlMessage) error {
	if c.conf.WriteTimeout > 0 {
		c.rawConn.SetWriteDeadline(time.Now().Add(c.conf.WriteTimeout))
	} else {
		// A deadline of zero means never timeout
		// The initial Time struct is zero
		c.rawConn.SetWriteDeadline(time.Time{})
	}
	return c.rawConn.WriteTo(h, p, cm)
}

// Creates the ip header message
func createIPHeader(sip, dip [4]byte) ipv4.Header {
	return ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 40,
		FragOff:  0,
		TTL:      64,
		Protocol: 6,
		Src:      sip[:],
		Dst:      dip[:],
	}
}

// Creates the control message
func createCM(sip, dip [4]byte) ipv4.ControlMessage {
	return ipv4.ControlMessage{
		TTL:     64,
		Src:     sip[:],
		Dst:     dip[:],
		IfIndex: 0,
	}
}

func (c *Channel) createTcpHead(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, prevSequence uint32, maskIndex int) (ipv4.Header, layers.TCP, []byte, error) {
	// Based on a preliminary investigation of my machine (running Ubuntu 18.04),
	// SYN packets always seem to have a window of 65495
	tcph.Window = 65495

	// We must generate a random TCP sequence number.
	// Changes in sequence number are used in bounce mode to detect duplicate
	// SYN/ACK packets from the legitimate TCP server
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tcph.Seq = r.Uint32() & 0xFFFFFFFF

	var (
		newipv4h ipv4.Header
		newtcph  layers.TCP
		newbuf   []byte
		err 		 error
	)

	newipv4h, newtcph, newbuf, _, err = c.conf.Encoder.SetByte(ipv4h, tcph, buf, maskIndex)
	if tcph.Seq == prevSequence {
		tcph.Seq = r.Uint32() & 0xFFFFFFFF
		newipv4h, newtcph, newbuf, _, err = c.conf.Encoder.SetByte(ipv4h, tcph, buf, maskIndex)
	}
	return newipv4h, newtcph, newbuf, err
}

// Create the TCP header IP payload
// prevSequence is used to ensure that the next sequence number
// differs from the previously sent sequence number.
// Flags are passed using the tcph argument. Only flags should
// be set, as any other field may be overwritten to make the header
// better resemble normal traffic
// The data byte b will be added to the tcp header as specified by the
// Encoder in the configuration
func createTcpHeadBuf(tcph layers.TCP, sip, dip [4]byte, sport, dport uint16) ([]byte, error) {
	var err error

	iph := layers.IPv4{
		Version:  4,
		Length:   20,
		TTL:      64,
		Protocol: 6,
		SrcIP:    sip[:],
		DstIP:    dip[:],
	}

	tcph.SrcPort = layers.TCPPort(sport)
	tcph.DstPort = layers.TCPPort(dport)

	if err = tcph.SetNetworkLayerForChecksum(&iph); err != nil {
		return nil, err
	}

	sb := gopacket.NewSerializeBuffer()
	op := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// This will compute proper checksums
	if err = tcph.SerializeTo(sb, op); err != nil {
		return nil, err
	}

	return sb.Bytes(), nil
}
