package tcpSyn

import (
	"../embedders"
	"bytes"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
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
	Embedder  embedders.TcpIpEmbedder

	// Timeouts for reads and writes
	// These are the inter packet timeout, which will always translate to the inter byte Timeout
	// since at this time each packet may only contain one byte
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

// A TCP covert channel
type Channel struct {
	conf       Config
	rawConn    *ipv4.RawConn
	cancel     chan bool
	recvChan   chan embedders.TcpIpPacket
	closeMutex *sync.Mutex
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	c := &Channel{conf: conf, cancel: make(chan bool), recvChan: make(chan embedders.TcpIpPacket, 1024), closeMutex: &sync.Mutex{}}
	if c.conf.Embedder == nil {
		c.conf.Embedder = &embedders.TcpIpSeqEncoder{}
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

	var (
		saddr        [4]byte
		sport, dport uint16
	)

	// Figure out the expected source and destination IP address
	// These change depending on whether or not we are in bounce mode
	if c.conf.Bounce {
		saddr, sport, dport = c.conf.BounceIP, c.conf.BouncePort, c.conf.OriginPort
	} else {
		saddr, sport, dport = c.conf.FriendIP, c.conf.FriendPort, c.conf.OriginPort
	}

	go c.readLoop(saddr, sport, dport)

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
	dataBuf, err := embedders.GetBuf(c.conf.Embedder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	if len(dataBuf) == 0 && c.conf.Delimiter == Buffer {
		return 0, nil
	}

	var (
		prev_val uint32 = 0
		first    bool   = true
		// There is guaranteed to be at least one space for a byte in the
		// data buffer at this point
		pos   uint64 = 0
		fin   bool
		p     embedders.TcpIpPacket
		state embedders.State = embedders.MakeState(c.conf.Embedder.GetMask())
	)
readloop:
	for {
		p, fin, err = c.readPacket(func(p embedders.TcpIpPacket) (embedders.TcpIpPacket, bool, bool, error) {
			// Check if done
			if c.conf.Delimiter == Protocol {
				if (!c.conf.Bounce && p.Tcph.ACK && !p.Tcph.RST) || (c.conf.Bounce && p.Tcph.RST) {
					return p, true, true, nil
				}
			}

			if c.conf.Bounce {
				p.Tcph.Seq = p.Tcph.Ack - 1
			}

			if (c.conf.Bounce && p.Tcph.SYN && p.Tcph.ACK) || (!c.conf.Bounce && p.Tcph.SYN && !p.Tcph.ACK) {
				if first || p.Tcph.Seq != prev_val {
					return p, true, false, nil
				}
			}
			return p, false, false, nil
		})

		if err != nil || fin {
			break readloop
		} else {
			prev_val = p.Tcph.Seq
			first = false

			var newBytes []byte

			newBytes, state, err = c.conf.Embedder.GetByte(p, state)
			if err != nil {
				break readloop
			}

			state = state.IncrementState()

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
		}
	}

	return embedders.CopyData(c.conf.Embedder.GetMask(), pos, dataBuf, data, err)
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
		rem          []byte
		n            uint64 = 0
		payload      []byte
		cm           ipv4.ControlMessage
		wait         time.Duration
		p            embedders.TcpIpPacket
		// We make it clear that the error always starts as nil
		err   error           = nil
		state embedders.State = embedders.MakeState(c.conf.Embedder.GetMask())
	)

	data, err = embedders.EncodeFromMask(c.conf.Embedder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	// The source and destination depend on whether or not we are in bounce mode
	if c.conf.Bounce {
		saddr, daddr, sport, dport = c.conf.FriendIP, c.conf.BounceIP, c.conf.FriendPort, c.conf.BouncePort
	} else {
		saddr, daddr, sport, dport = c.conf.OriginIP, c.conf.FriendIP, c.conf.OriginPort, c.conf.FriendPort
	}

	rem = data

readloop:
	for len(rem) > 0 {
		p.Ipv4h = createIPHeader(saddr, daddr)
		cm = createCM(saddr, daddr)

		p.Tcph.SYN = true

		p, rem, wait, state, err = c.createTcpHead(p, rem, prevSequence, state)
		if err != nil {
			break readloop
		}
		prevSequence = p.Tcph.Seq

		payload, err = createTcpHeadBuf(p.Tcph, saddr, daddr, sport, dport)
		if err != nil {
			break readloop
		}

		// We wait for the time specified by the user's GetDelay function
		// or until the user signals to cancel
		select {
		case <-time.After(wait):
		case <-c.cancel:
			err = &WriteWaitCancel{}
			break readloop
		}

		if err = c.writeConn(&p.Ipv4h, payload, &cm); err != nil {
			break readloop
		}
		state = state.IncrementState()
		n = uint64(len(data) - len(rem))
	}

	// Readjust size to represent number of bytes sent
	n, err = embedders.GetSentSize(c.conf.Embedder.GetMask(), n, err)
	if err != nil {
		return n, err
	}

	// If we are using Protocol delimiting, we must send a final
	// ACK packet to alert the receiver that the message is done sending
	if c.conf.Delimiter == Protocol {
		p.Ipv4h = createIPHeader(saddr, daddr)
		cm = createCM(saddr, daddr)

		// We want to still encode data so that this final packet looks
		// like others. To do that, we need to provide enough bytes
		// to fit the mask at this maskIndex
		// We create a buffer with the appropriate size and fill it with
		// random numbers
		var fakeBuf []byte = make([]byte, len(c.conf.Embedder.GetMask()[state.MaskIndex]))
		for i := range fakeBuf {
			fakeBuf[i] = byte(r.Uint32())
		}

		p.Tcph = layers.TCP{ACK: true}
		p, _, _, state, err = c.createTcpHead(p, fakeBuf, prevSequence, state)
		if err != nil {
			return n, err
		}
		prevSequence = p.Tcph.Seq

		payload, err = createTcpHeadBuf(p.Tcph, saddr, daddr, sport, dport)
		if err != nil {
			return n, err
		}

		if err = c.writeConn(&p.Ipv4h, payload, &cm); err != nil {
			return n, err
		}
	}

	return n, nil
}

// Closes the covert channel.
// If the covert channel is closed while a read or write is occuring then
// the function function will return with an OpCancel error
func (c *Channel) Close() error {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()
	select {
	case <-c.cancel:
	default:
		close(c.cancel)
	}
	return c.rawConn.Close()
}

// Read from a raw connection whil setting a timeout if necessary
func (c *Channel) readPacket(f func(p embedders.TcpIpPacket) (embedders.TcpIpPacket, bool, bool, error)) (embedders.TcpIpPacket, bool, error) {
	var (
		p         embedders.TcpIpPacket
		err       error
		valid     bool
		fin       bool
		startTime time.Time = time.Now()
	)
	for {
		if c.conf.ReadTimeout > 0 {
			select {
			case p = <-c.recvChan:
				p, valid, fin, err = f(p)
				if valid || err != nil {
					return p, fin, err
				} else if time.Now().Sub(startTime) > c.conf.ReadTimeout {
					return p, fin, errors.New("Read Timeout")
				}
			case <-time.After(c.conf.ReadTimeout):
				return p, fin, errors.New("Read Timeout")
			case <-c.cancel:
				return p, fin, errors.New("Cancel")
			}
		} else {
			select {
			case p = <-c.recvChan:
				p, valid, fin, err = f(p)
				if valid || err != nil {
					return p, fin, err
				}
			case <-c.cancel:
				return p, fin, errors.New("Cancel")
			}
		}
	}
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

func (c *Channel) createTcpHead(p embedders.TcpIpPacket, buf []byte, prevSequence uint32, state embedders.State) (embedders.TcpIpPacket, []byte, time.Duration, embedders.State, error) {
	// Based on a preliminary investigation of my machine (running Ubuntu 18.04),
	// SYN packets always seem to have a window of 65495
	p.Tcph.Window = 65495

	// We must generate a random TCP sequence number.
	// Changes in sequence number are used in bounce mode to detect duplicate
	// SYN/ACK packets from the legitimate TCP server
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var (
		newpkt embedders.TcpIpPacket
		newbuf []byte
		err    error
		wait   time.Duration
	)

	p.Tcph.Seq = r.Uint32() & 0xFFFFFFFF
	for p.Tcph.Seq == prevSequence {
		p.Tcph.Seq = r.Uint32() & 0xFFFFFFFF
	}
	newpkt, newbuf, wait, state, err = c.conf.Embedder.SetByte(p, buf, state)
	if newpkt.Tcph.Seq == prevSequence {
		return newpkt, newbuf, time.Duration(0), state, errors.New("Could not send packet; " +
			"Embedder set current sequence number to preceeding sequence number. " +
			"Would prevent receiver from differentiating packets")
	}
	return newpkt, newbuf, wait, state, err
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

// A loop that continuously receives packets across the raw socket
// Incoming packets are analysed to confirm that they
// have the expected source IP address (Our friend's IP address)
func (c *Channel) readLoop(saddr [4]byte, sport, dport uint16) {
	var (
		buf [1024]byte
		l   *log.Logger = log.New(os.Stderr, "", log.Flags())
	)
	for {
		h, p, _, err := c.rawConn.ReadFrom(buf[:])

		// Once the covert channel is closed we
		// must exit this read loop
		select {
		case <-c.cancel:
			return
		default:
			if err != nil {
				continue
			}
		}

		tcph := layers.TCP{}
		if err = tcph.DecodeFromBytes(p, gopacket.NilDecodeFeedback); err == nil {
			if bytes.Equal(h.Src.To4(), saddr[:]) {

				// When reading options DecodeFromBytes does not copy option data,
				// it merely slices into the array. We copy here to make sure that the data
				// does not get overwritten when the underlying array is used for later reads
				for i := range tcph.Options {
					tcph.Options[i].OptionData = append([]byte{}, tcph.Options[i].OptionData...)
				}

				if tcph.SrcPort == layers.TCPPort(sport) && tcph.DstPort == layers.TCPPort(dport) {
					select {
					case c.recvChan <- embedders.TcpIpPacket{Ipv4h: *h, Tcph: tcph, Time: time.Now()}:
					default:
						l.Println("Too many packets: Dropped Packet")
					}
				}
			}
		}
	}
}
