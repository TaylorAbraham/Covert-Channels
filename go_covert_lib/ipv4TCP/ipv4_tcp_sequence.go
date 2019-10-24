package ipv4TCP

import (
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
	Protocol = 1
)

type WriteWaitCancel struct {}

func (o *WriteWaitCancel) Error() string {
	return "Write Cancelled"
}

type Config struct {
	// The configuration of an ipv4_tcp_sequence covert channel
	// This structure recognizes three IP-port pairs, the friend, the origin, and an optional bounce.
	// The friend is the node you are sending messages to.
	// The origin is where messages are send when they reach you.
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
	Bounce     bool
	// The delimiter to use to deliniate messages. Currently it is either no deliniation (Delim::None)
	// or delinieating by a TCP packet with a specific flag (Delim::Protocol).
	// Default is Delim::Protocol.
	Delimiter  uint8
	Encoder    TcpEncoder
	// A function to retrieve a delay to implement between sent packets. By default this
	// function returns a delay of 0 ms, but users can set it to a longer time or even to
	// their favourite distribution.
	GetDelay   func() time.Duration

	// Timeouts for reads and writes
	// These are the inter packet timeout, which will always translate to the inter byte Timeout
	// since at this time each packet may only contain one byte
	WriteTimeout time.Duration
	ReadTimeout time.Duration
}

// An encoded may be provided to the TCP Channel
// to configure where the covert message is hidden
// A method is used to hide the byte during sending, and a second is used
// extract it during reception
type TcpEncoder interface {
	// We use the prevSequence to indicate duplicate packets arriving
	// on from the bouncer socket. It is this function's responsibility
	// to ensure that any modifications to the tcp header ensure that the
	// new sequence number is different than the previos sequence number
	SetByte(tcph layers.TCP, b byte, prevSequence uint32) (layers.TCP, error)
	GetByte(tcph layers.TCP, bounce bool) (byte, error)
}

// The default TcpEncoder that hides the covert message in the
// sequence number
type SeqEncoder struct{}

// Since this function explicitely modifies the sequence number, we must ensure
// we generate a random one in the same way as createPacket
// Other implementations of this function may use other headers, and so would
// not be required to duplicate this
func (s *SeqEncoder) SetByte(tcph layers.TCP, b byte, prevSequence uint32) (layers.TCP, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tcph.Seq = (r.Uint32() & 0xFFFFFF00) | uint32(b)
	// We hide the byte in the low 8 bits of the randomly generated sequence number
	for tcph.Seq == prevSequence {
		tcph.Seq = (r.Uint32() & 0xFFFFFF00) | uint32(b)
	}
	return tcph, nil
}

// Retrieve the byte from a packet encoded by the sequence number Encoder
// If this covert channel was using the bounce functionality then the
// value is moved to the acknowledgement number by the legitimate TCP socket
func (s *SeqEncoder) GetByte(tcph layers.TCP, bounce bool) (byte, error) {
	if bounce {
		return byte((0xFF & tcph.Ack) - 1), nil
	} else {
		return byte(0xFF & tcph.Seq), nil
	}
}

// A TCP covert channel
type Channel struct {
	conf Config
	rawConn *ipv4.RawConn
	writeCancel chan bool
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	c := &Channel{conf : conf, writeCancel : make(chan bool)}
	if c.conf.Encoder == nil {
		c.conf.Encoder = &SeqEncoder{}
	}

	conn, err := net.ListenPacket("ip4:6", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	c.rawConn, err = ipv4.NewRawConn(conn)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Receive a covert message
// data is a buffer for the message.
// If the channel is in protocol delimiter mode then this function
// will return as soon as it receives the proper delimiter.
// Otherwise, it will wait until the buffer is filled.
// The optional progress chan has not yet been implemented.
// We return the number of bytes received even if an error is encountered,
// in which case data will have valid received bytes up to that point.
func (c *Channel) Receive(data []byte, progress chan<- uint64) (uint64, error) {

	if len(data) == 0 {
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
	)

	// Figure out the expected source and destination IP address
	// These change depending on whether or not we are in bounce mode
	if c.conf.Bounce {
		saddr, sport, dport = c.conf.BounceIP, c.conf.BouncePort, c.conf.OriginPort
	} else {
		saddr, sport, dport = c.conf.FriendIP, c.conf.FriendPort, c.conf.OriginPort
	}

	for {
		h, p, _, err := c.readConn(buf)
		if err != nil {
			return pos, err
		}
		tcph := layers.TCP{}
		tcph.DecodeFromBytes(p, gopacket.NilDecodeFeedback)
		// We check for the expected source IP, source port, and destination port
		if bytes.Equal(h.Src.To4(), saddr[:]) {
			if tcph.SrcPort == layers.TCPPort(sport) && tcph.DstPort == layers.TCPPort(dport) {
				if c.conf.Delimiter == Protocol {
					if (!c.conf.Bounce && tcph.ACK) || (c.conf.Bounce && tcph.RST) {
						return pos, nil
					}
				}
				if c.conf.Bounce {
					prev_val = tcph.Ack
				} else {
					prev_val = tcph.Seq
				}

				// Check the expected flags
				if (c.conf.Bounce && tcph.SYN && tcph.ACK) || (!c.conf.Bounce && tcph.SYN && !tcph.ACK) {
					if first || current_val != prev_val {
						prev_val = current_val
						first = false

						b, err := c.conf.Encoder.GetByte(tcph, c.conf.Bounce)
						if err != nil {
							return pos, err
						}
						// Make sure we don't overflow the buffer
						// The buffer can only overflow if we are in protocol delimiter mode
						// Once the buffer fills in protocol delimiter mode we
						// wait for the end packet. If a packet arrives that does not
						// have the expected flags then we will reach this statement
						// and be unable to put the new byte into the buffer
						// In that case we return with an error indicating buffer overflow
						// (see below)
						if pos < uint64(len(data)) {
							data[pos] = b
						} else if pos == uint64(len(data)) && c.conf.Delimiter == Protocol {
							// If we are using the protocol delimiter, then we can fill the full
							// buffer and then wait for the end packet. It is only if more bytes
							// are received that we must notify of an error
							return pos, errors.New("End packet not received, buffer full")
						}
						pos += 1
						// We have filled the buffer without protocol delimiter and we should return immediately
						if pos == uint64(len(data)) && c.conf.Delimiter != Protocol {
							return pos, nil
						}
					}
				}
			}
		}
	}
}

// Send a covert message
// data is the entire message that will be sent.
// The optional progress chan can be used to alert the user as to the progress
// of message transmission. This is useful if the user has set the GetDelay
// function for the channel. The channel should be buffered, otherwise the
// update will be skipped if it is not immediately read.
// The GetDelay function can be used to set a large
// inter packet delay to help obscure the communication. In that case
// the progress channel will fire whenever the number of sent bytes has risen
// by at least 1 percent.
// We return the number of bytes sent even if an error is encountered
func (c *Channel) Send(data []byte, progress chan<- uint64) (uint64, error) {
	// We make it clear that the error always starts as nil
	var err error = nil

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var (
		prevSequence uint32 = r.Uint32()
		saddr, daddr [4]byte
		sport, dport uint16
		num          uint64 = 0
		h            *ipv4.Header
		p            []byte
		cm           *ipv4.ControlMessage
		wait         time.Duration
		sendPercent  uint64 = 0
	)

	// The source and destination depend on whether or not we are in bounce mode
	if c.conf.Bounce {
		saddr, daddr, sport, dport = c.conf.FriendIP, c.conf.BounceIP, c.conf.FriendPort, c.conf.BouncePort
	} else {
		saddr, daddr, sport, dport = c.conf.OriginIP, c.conf.FriendIP, c.conf.OriginPort, c.conf.FriendPort
	}

	for _, b := range data {
		h = c.createIPHeader(saddr, daddr)
		cm = c.createCM(saddr, daddr)

		p, prevSequence, err = c.createTcpHeadBuf(b, prevSequence, layers.TCP{SYN: true}, saddr, daddr, sport, dport)
		if err != nil {
			return num, err
		}

		if err = c.writeConn(h, p, cm); err != nil {
			return num, err
		}

		if progress != nil {
			var currPercent uint64 = uint64(float64(num) / float64(len(data)))
			if currPercent > sendPercent {
				sendPercent = currPercent
				// We send the update along the channel
				// To avoid deadlock we skip if the channel
				// is not ready to receive.
				// As such, the channel should be buffered.
				select {
					case progress <- currPercent:
					default:
				}
			}
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
			return num, &WriteWaitCancel{}
		}
	}
	// If we are using Protocol delimiting, we must send a final
	// ACK packet to alert the receiver that the message is done sending
	if c.conf.Delimiter == Protocol {
		h = c.createIPHeader(saddr, daddr)
		cm = c.createCM(saddr, daddr)

		p, prevSequence, err = c.createTcpHeadBuf(uint8(r.Uint32()), prevSequence, layers.TCP{ACK: true}, saddr, daddr, sport, dport)
		if err != nil {
			return num, err
		}

		if err = c.writeConn(h, p, cm); err != nil {
			return num, err
		}
	}

	return num, nil
}

// Creates the ip header message
func (c *Channel) createIPHeader(sip, dip [4]byte) *ipv4.Header {
	return &ipv4.Header{
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
func (c *Channel) createCM(sip, dip [4]byte) *ipv4.ControlMessage {
	return &ipv4.ControlMessage{
		TTL:     64,
		Src:     sip[:],
		Dst:     dip[:],
		IfIndex: 0,
	}
}

// Closes the covert channel.
// If the covert channel is closed while a read or write is occuring then
// the function function will return with an OpCancel error
func (c *Channel) Close() error {
	// The write operation allows the user to specify
	// a delay between packets
	// We can't just rely on the raw connection being
	// close, we must also be able to cancel this delay,
	// which is done with the writeCancel method
	select {
		case <-c.writeCancel:
		default:
			close(c.writeCancel)
	}
	return c.rawConn.Close()
}

// Read to a raw connection whil setting a timeout if necessary
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

// Write to a raw connection whil setting a timeout if necessary
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

// Create the TCP header IP payload
// prevSequence is used to ensure that the next sequence number
// differs from the previously sent sequence number.
// Flags are passed using the tcph argument. Only flags should
// be set, as any other field may be overwritten to make the header
// better resemble normal traffic
// The data byte b will be added to the tcp header as specified by the
// Encoder in the configuration
func (c *Channel) createTcpHeadBuf(b byte, prevSequence uint32, tcph layers.TCP, sip, dip [4]byte, sport, dport uint16) ([]byte, uint32, error) {
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
	// Based on a preliminary investigation of my machine (running Ubuntu 18.04),
	// SYN packets always seem to have a window of 65495
	tcph.Window = 65495

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tcph.Seq = r.Uint32() & 0xFFFFFFFF
	for tcph.Seq == prevSequence {
		tcph.Seq = r.Uint32() & 0xFFFFFFFF
	}
	tcph, err = c.conf.Encoder.SetByte(tcph, b, prevSequence)
	if err != nil {
		return nil, prevSequence, err
	}

	if err := tcph.SetNetworkLayerForChecksum(&iph); err != nil {
		return nil, prevSequence, err
	}

	sb := gopacket.NewSerializeBuffer()
	op := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// This will compute proper checksums
	if err := tcph.SerializeTo(sb, op); err != nil {
		return nil, prevSequence, err
	}

	return sb.Bytes(), tcph.Seq, nil
}
