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
	"sync"
)

const (
	Protocol = 1
)

type OpCancel struct {}

func (o *OpCancel) Error() string {
	return "Operation Cancelled"
}

type Config struct {
	FriendIP   [4]byte
	OriginIP   [4]byte
	FriendPort uint16
	OriginPort uint16
	Bounce     bool
	Delimiter  uint8
	Encoder    TcpEncoder
	GetDelay   func() time.Duration
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
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	c := &Channel{conf : conf}
	if c.conf.Encoder == nil {
		c.conf.Encoder = &SeqEncoder{}
	}
	return c, nil
}

// Receive a covert message
// data is a buffer for the message.
// If the channel is in protocol delimiter mode then this function
// will return as soon as it receives the proper delimiter.
// Otherwise, it will wait until the buffer is filled.
// The optional progress chan has not yet been implemented.
// The cancel chan will cancel message reception, causing an early return
// regardless of delimiter mode. If the reception is cancelled an OpCancel error
// is returned. The caller should close this
// channel to cancel the transmission or else risk deadlock.
// We return the number of bytes received even if an error is encountered,
// in which case data will have valid received bytes up to that point.
func (c *Channel) Receive(data []byte, progress chan<- uint64, cancel <-chan struct{}) (uint64, error) {
	return c.receive(data, progress, cancel, nil)
}

// Same as Receive, but with an added channel parameter to indicate when the raw
// socket has been opened and the channel is ready to receive
// The ready channel is closed when the raw socket has been opened
// Therefore a new channel must be provided for every call to receive
// This is primarily used for testing, where it is necessary to
// coordinate the send to occur after the receive is ready.
// It is assumed that in a real situation the sender and receiver are on different machines,
// so that the receiver would have to prepare to receive well before the message is actually
// send (i.e. the receiver can afford a few milliseconds delay in setting up the read)
func (c *Channel) receive(data []byte, progress chan<- uint64, cancel <-chan struct{}, ready chan<- struct{}) (uint64, error) {
	var once sync.Once
	readyDone := func () {
		// Indicates that the raw socket has been opened and the channel is ready to read
		if ready != nil {
			close(ready)
		}
	}
	// readyDone is called either when we return or once the raw socket is opened,
	// whichever comes first
	// It must be called even if there is an error
	defer	once.Do(readyDone)

	if len(data) == 0 {
		return 0, nil
	}

	conn, err := net.ListenPacket("ip4:6", "0.0.0.0")
	if err != nil {
		return 0, err
	}

	// Setup to allow closing waiting connections
	var closer chan struct{} = make(chan struct{})
	var done chan struct{} = make(chan struct{})

	// This goroutine will wait for either the message being fully Received
	// or the user sending a message on the cancel channel
	// It is setup to allow shutdown whant the function returns
	go func() {
		if cancel == nil {
			<-closer
		} else {
			select {
			case <-cancel:
			case <-closer:
			}
		}
		conn.Close()
		close(done)
	}()

	// close the connection when the function returns
	// (if it hasn't already been closed
	// and wait for the above goroutine to return
	defer func() {
		close(closer)
		<-done
	}()

	raw, err := ipv4.NewRawConn(conn)
	if err != nil {
		return 0, err
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
		saddr, sport, dport = c.conf.OriginIP, c.conf.OriginPort, c.conf.FriendPort
	} else {
		saddr, sport, dport = c.conf.FriendIP, c.conf.FriendPort, c.conf.OriginPort
	}

	once.Do(readyDone)

	for {
		h, p, _, err := raw.ReadFrom(buf)
		if err != nil {
			return pos, err
			// Cancelling the read by closing the conn causes an errors
			// Unfortunately there does not seem to be a way to use the socket shutdown
			// function on the raw sockets provided by this library,
			// as is the case with the Rust Socket2 library
		} else if len(p) == 0 {
			return pos, &OpCancel{}
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
// The cancel chan will cancel transmission, which can be useful if the user
// wants to cancel a long running transmission. The caller should close this
// channel to cancel the transmission or else risk deadlock. If the Write
// is cancelled and Protocol delimiting is active, the delimiter packet is
// transmitted. When cancelling this function will return with an OpCancel error
// unless an error is encountered when transmitting an delimiter packets, in which
// case that error is returned.
// We return the number of bytes sent even if an error is encountered
func (c *Channel) Send(data []byte, progress chan<- uint64, cancel <-chan struct{}) (uint64, error) {
	// We make it clear that the error always starts as nil
	var err error = nil

	conn, err := net.ListenPacket("ip4:6", "0.0.0.0")
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	raw, err := ipv4.NewRawConn(conn)
	if err != nil {
		return 0, err
	}

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
		saddr, daddr, sport, dport = c.conf.FriendIP, c.conf.OriginIP, c.conf.FriendPort, c.conf.OriginPort
	} else {
		saddr, daddr, sport, dport = c.conf.OriginIP, c.conf.FriendIP, c.conf.OriginPort, c.conf.FriendPort
	}

loop:
	for _, b := range data {
		h = c.createIPHeader(saddr, daddr)
		cm = c.createCM(saddr, daddr)

		p, prevSequence, err = c.createTcpHeadBuf(b, prevSequence, layers.TCP{SYN: true}, saddr, daddr, sport, dport)
		if err != nil {
			return num, err
		}

		if err = raw.WriteTo(h, p, cm); err != nil {
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

		// We wait for the time specified by the user's GetDelay function
		// or until the user signals to cancel
		select {
		case <-time.After(wait):
		case <-cancel:
			// We don't return immediately
			// so that we send the protocol delimiter packet
			// even if the write has been cancelled
			// An error in creating or writing the delimiter packet
			// will take precedence over the "Write Cancelled" error
			err = &OpCancel{}
			break loop
		}
		num += 1
	}
	// If we are using Protocol delimiting, we must send a final
	// ACK packet to alert the receiver that the message is done sending
	if c.conf.Delimiter == Protocol {
		var errDelim error = nil
		h = c.createIPHeader(saddr, daddr)
		cm = c.createCM(saddr, daddr)

		p, prevSequence, errDelim = c.createTcpHeadBuf(uint8(r.Uint32()), prevSequence, layers.TCP{ACK: true}, saddr, daddr, sport, dport)
		if errDelim != nil {
			return num, errDelim
		}

		if errDelim = raw.WriteTo(h, p, cm); errDelim != nil {
			return num, errDelim
		}
	}

	return num, err
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
