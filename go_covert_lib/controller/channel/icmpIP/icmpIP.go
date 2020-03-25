package icmpIP

import (
	//"../embedders"
	//"bytes"
	//"context"
	//"errors"
	//"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
	"time"
)

// We make the fields public for logging
type packet struct {
	Ipv4h ipv4.Header
	icmph  layers.ICMPv4
}

type syncPktMap struct {
	mutex  *sync.Mutex
	pktMap map[uint16][]packet
}

type acceptedConn struct {
	conn net.Conn
	friendPort uint16
}

type Config struct {
	FriendIP          [4]byte
	OriginIP          [4]byte
	FriendReceivePort uint16
	OriginReceivePort uint16
	DialTimeout       time.Duration
	Encoder           IcmpEncoder

	AcceptTimeout time.Duration

	// The intra-packet read timeout. Set zero for no timeout.
	// The receive method will block until a three way handshake
	// is complete and the listener returns, and will only exit with a
	// read timeout if the intra-packet delay is too large.
	ReadTimeout time.Duration
	// The timeout for writing the packet to a raw socket. Set zero for no timeout.
	WriteTimeout time.Duration
}

type Channel struct {
	conf    Config
	rawConn *ipv4.RawConn
	cancel  chan bool

	// For debugging purposes, log all packets received and sent
	sendPktLog    *syncPktMap
	receivePktLog *syncPktMap

	// We make the mutex a pointer to avoid the risk of copying
	writeMutex *sync.Mutex
	closeMutex *sync.Mutex

	acceptChan chan acceptedConn
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
		conf:          conf,
		cancel:        make(chan bool),
		writeMutex:    &sync.Mutex{},
		closeMutex:    &sync.Mutex{},
	}

	if c.conf.Encoder == nil {
		c.conf.Encoder = &IDEncoder{}
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
	return 0, nil
}

func (c *Channel) Receive(data []byte) (uint64, error) {
	return 0, nil
}

func (c *Channel) sendPacket(h *ipv4.Header, b []byte, cm *ipv4.ControlMessage) error {
	return nil
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
		Protocol: 17,
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
