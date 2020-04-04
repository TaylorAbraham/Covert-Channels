package icmpIP

import (
	"../embedders"
	"bytes"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
	"time"
)

type Config struct {
	FriendIP [4]byte
	OriginIP [4]byte
	Embedder IcmpEncoder

	// The intra-packet read timeout. Set zero for no timeout.
	// The receive method will block until a three way handshake
	// is complete and the listener returns, and will only exit with a
	// read timeout if the intra-packet delay is too large.
	ReadTimeout time.Duration
	// The timeout for writing the packet to a raw socket. Set zero for no timeout.
	WriteTimeout time.Duration
	Identifier   uint16
}

type Channel struct {
	conf    Config
	rawConn *ipv4.RawConn
	cancel  chan bool

	// We make the mutex a pointer to avoid the risk of copying
	writeMutex *sync.Mutex
	closeMutex *sync.Mutex
}

func (c *Channel) Close() error {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()
	select {
	// Have we already closed
	case <-c.cancel:
		return nil
	default:
		close(c.cancel)
	}
	return c.rawConn.Close()
}

//create channel
func MakeChannel(conf Config) (*Channel, error) {

	c := &Channel{
		conf:       conf,
		cancel:     make(chan bool),
		writeMutex: &sync.Mutex{},
		closeMutex: &sync.Mutex{},
	}

	if c.conf.Embedder == nil {
		c.conf.Embedder = &IDEncoder{}
	}

	//ip network within the ICMP protocol
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
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

func (c *Channel) Send(data []byte) (uint64, error) {
	data, err := embedders.EncodeFromMask(c.conf.Embedder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	var (
		ipv4h ipv4.Header         = createIPHeader(c.conf.OriginIP, c.conf.FriendIP)
		cm    ipv4.ControlMessage = createCM(c.conf.OriginIP, c.conf.FriendIP)
		icmph layers.ICMPv4
		wbuf  []byte
		rem   []byte = data
		n     uint64
		state embedders.State = embedders.MakeState(c.conf.Embedder.GetMask())
	)

	// Send each packet
	for len(rem) > 0 {

		var payload []byte = make([]byte, 26) //set payload of length 26
		if ipv4h, icmph, rem, state, err = c.conf.Embedder.SetByte(ipv4h, icmph, rem, state); err != nil {
			break
		}

		if wbuf, icmph, err = createicmpheader(icmph, payload, c.conf.Identifier); err != nil {
			break
		}

		if err = c.sendPacket(&ipv4h, wbuf, &cm); err != nil {
			break
		}
		state = state.IncrementState()
		n = uint64(len(data) - len(rem))
	}

	// Readjust size to represent number of bytes sent
	n, err = embedders.GetSentSize(c.conf.Embedder.GetMask(), n, err)
	if err != nil {
		return n, err
	}

	var payload []byte = make([]byte, 0) //length 0 signifies the end of the message

	if wbuf, icmph, err = createicmpheader(icmph, payload, c.conf.Identifier); err != nil {
		return n, err
	}

	err = c.sendPacket(&ipv4h, wbuf, &cm)
	return n, nil
}

func createicmpheader(icmph layers.ICMPv4, payload []byte, identifier uint16) ([]byte, layers.ICMPv4, error) {
	// assigning type code to the ICMP layer
	icmph.TypeCode = layers.CreateICMPv4TypeCode(1, 0)
	icmph.Id = identifier

	sb := gopacket.NewSerializeBuffer()
	op := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	if err := gopacket.SerializeLayers(sb, op, &icmph, gopacket.Payload(payload)); err != nil {
		return nil, icmph, err
	}

	return sb.Bytes(), icmph, nil
}

func (c *Channel) Receive(data []byte) (uint64, error) {

	// We must expand out the input storage array to
	// the correct size to potentially handle variable size inputs
	dataBuf, err := embedders.GetBuf(c.conf.Embedder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	var (
		buf   []byte = make([]byte, 1024)
		saddr [4]byte
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
		h              *ipv4.Header
		p              []byte
		state          embedders.State = embedders.MakeState(c.conf.Embedder.GetMask())
	)

	saddr = c.conf.FriendIP

	prevPacketTime = time.Now()

	for {
		h, p, _, err = c.readConn(buf)

		if err != nil {
			break
		}
		icmph := layers.ICMPv4{}
		if err = icmph.DecodeFromBytes(p, gopacket.NilDecodeFeedback); err == nil {
			// We check for the expected source IP, source port, and destination port
			if bytes.Equal(h.Src.To4(), saddr[:]) {
				if icmph.Id == c.conf.Identifier && icmph.TypeCode == layers.CreateICMPv4TypeCode(1, 0) {
					if len(p) == 8 { //end of message
						break
					} else { //the rest of the message
						prevPacketTime = time.Now()
						var b []byte
						b, state, err = c.conf.Embedder.GetByte(*h, icmph, state)
						if err != nil {
							break
						}
						state = state.IncrementState()
						if (pos + uint64(len(b))) >= uint64(len(dataBuf)) { //overflow
							err = errors.New("Overflow. End of message never received.")
							break
						} else { //add data to array, increment pos
							for i := 0; i < len(b); i++ {
								dataBuf[pos] = b[i]
								pos++
							}
							prevPacketTime = time.Now()
						}
					}
				}
			}
		}
		if c.conf.ReadTimeout > 0 && time.Now().Sub(prevPacketTime) > c.conf.ReadTimeout {
			err = errors.New("Covert Packet Timeout")
			break
		}
	}
	return embedders.CopyData(c.conf.Embedder.GetMask(), pos, dataBuf, data, err)
}

func (c *Channel) sendPacket(h *ipv4.Header, b []byte, cm *ipv4.ControlMessage) error {
	if c.conf.WriteTimeout > 0 {
		c.rawConn.SetWriteDeadline(time.Now().Add(c.conf.WriteTimeout))
	} else {
		// A deadline of zero means never timeout
		// The initial Time struct is zero
		c.rawConn.SetWriteDeadline(time.Time{})
	}
	return c.rawConn.WriteTo(h, b, cm)
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

// Creates the ip header message
func createIPHeader(sip, dip [4]byte) ipv4.Header {
	return ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		FragOff:  0,
		TTL:      64,
		Protocol: 1,
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
