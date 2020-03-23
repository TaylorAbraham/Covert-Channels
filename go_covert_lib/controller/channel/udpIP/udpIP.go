package udpIP

import (
	"../embedders"
	"bytes"
	"context"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
	"time"
)

const maxAccept = 32

// We make the fields public to facilitate logging
type packet struct {
	Ipv4h ipv4.Header
	Udph  layers.UDP
}

type syncPktMap struct {
	mutex  *sync.Mutex
	pktMap map[uint16][]packet
}

func MakeSyncMap() *syncPktMap {
	return &syncPktMap{pktMap: make(map[uint16][]packet), mutex: &sync.Mutex{}}
}

func (smap *syncPktMap) Add(k uint16, v packet) {
	smap.mutex.Lock()
	smap.pktMap[k] = append(smap.pktMap[k], v)
	smap.mutex.Unlock()
}

type acceptedConn struct {
	conn net.Conn
	// The UDP port used by our Friend IP in this covert message
	friendPort uint16
}

type Config struct {
	FriendIP          [4]byte
	OriginIP          [4]byte
	FriendReceivePort uint16
	OriginReceivePort uint16
	DialTimeout       time.Duration
	Encoder           UdpEncoder

	// For debugging purposes, log all packets that are sent or received
	logPackets bool

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
		sendPktLog:    MakeSyncMap(),
		receivePktLog: MakeSyncMap(),
		writeMutex:    &sync.Mutex{},
		closeMutex:    &sync.Mutex{},
		// Only 32 connections can be accepted before they begin to be dropped
		acceptChan: make(chan acceptedConn, maxAccept),
	}

	if c.conf.Encoder == nil {
		c.conf.Encoder = &IDEncoder{}
	}

	//ip network with udp protocol
	conn, err := net.ListenPacket("ip4:17", "0.0.0.0")
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

	data, err := embedders.EncodeFromMask(c.conf.Encoder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	nd := net.Dialer{
		Timeout: c.conf.DialTimeout,
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	// A channel to be closed after message sending is over
	// to end the goroutine used for cancelling the dial
	doneMsg := make(chan byte)
	// If the cancel method is called for the covert channel, we want to
	// exit the dial operation. This goroutine waits for the cancel channel
	// to be closed and cancels the context if so
	go func() {
		select {
		case <-c.cancel:
			cancelFn()
		// doneDial ensures that this go routine always exits
		case <-doneMsg:
		}
	}()
	// we close the doneMsg channel to ensure that the go routine always exits
	defer close(doneMsg)

	// DialContext
	conn, err := nd.DialContext(ctx, "udp4", (&net.UDPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.FriendReceivePort)}).String())
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	var (
		ipv4h ipv4.Header         = createIPHeader(c.conf.OriginIP, c.conf.FriendIP)
		cm    ipv4.ControlMessage = createCM(c.conf.OriginIP, c.conf.FriendIP)
		udph  layers.UDP
		wbuf  []byte
		rem   []byte = data
		n     uint64
		maskIndex int = 0
	)

	// Send each packet
	for len(rem) > 0 {
		var payload []byte = make([]byte, 26) //set payload of length 26
		if ipv4h, udph, rem, err = c.conf.Encoder.SetByte(ipv4h, udph, rem, maskIndex); err != nil {
			break
		}
		maskIndex = embedders.UpdateMaskIndex(c.conf.Encoder.GetMask(), maskIndex)

		if wbuf, udph, err = createUDPHeader(udph, c.conf.OriginIP, c.conf.FriendIP, c.conf.OriginReceivePort, c.conf.FriendReceivePort, payload); err != nil {
			break
		}

		if c.conf.logPackets {
			c.sendPktLog.Add(c.conf.OriginReceivePort, packet{Ipv4h: ipv4h, Udph: udph})
		}

		if err = c.sendPacket(&ipv4h, wbuf, &cm); err != nil {
			break
		}
		n = uint64(len(data) - len(rem))
	}

	// Readjust size to represent number of bytes sent
	n, err = embedders.GetSentSize(c.conf.Encoder.GetMask(), n, err)
	if err != nil {
		return n, err
	}

	//last payload determining the end of the message
	var payload []byte = make([]byte, 24) //length 24 signifies the end of the message

	if wbuf, udph, err = createUDPHeader(udph, c.conf.OriginIP, c.conf.FriendIP, c.conf.OriginReceivePort, c.conf.FriendReceivePort, payload); err != nil {
		return n, err
	}

	if c.conf.logPackets {
		c.sendPktLog.Add(c.conf.OriginReceivePort, packet{Ipv4h: ipv4h, Udph: udph})
	}

	err = c.sendPacket(&ipv4h, wbuf, &cm)
	return n, err
}

func (c *Channel) sendPacket(h *ipv4.Header, b []byte, cm *ipv4.ControlMessage) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	if c.conf.WriteTimeout != 0 {
		c.rawConn.SetWriteDeadline(time.Now().Add(c.conf.WriteTimeout))
	}
	return c.rawConn.WriteTo(h, b, cm)
}

// We return the tcph header so that it can be logged if needed for debugging
func createUDPHeader(udph layers.UDP, sip, dip [4]byte, sport, dport uint16, payload []byte) ([]byte, layers.UDP, error) {

	iph := layers.IPv4{
		SrcIP: sip[:],
		DstIP: dip[:],
	}

	udph.SrcPort = layers.UDPPort(sport)
	udph.DstPort = layers.UDPPort(dport)

	//udph.Length = uint16(len(payload))

	if err := udph.SetNetworkLayerForChecksum(&iph); err != nil {
		return nil, udph, err
	}

	sb := gopacket.NewSerializeBuffer()
	op := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// This will compute proper checksums
	if err := gopacket.SerializeLayers(sb, op, &udph, gopacket.Payload(payload)); err != nil {
		return nil, udph, err
	}

	return sb.Bytes(), udph, nil
}

func (c *Channel) Receive(data []byte) (uint64, error) {

	// We must expand out the input storage array to
	// the correct size to potentially handle variable size inputs
	dataBuf, err := embedders.GetBuf(c.conf.Encoder.GetMask(), data)
	if err != nil {
		return 0, err
	}

	if len(dataBuf) == 0 {
		return 0, nil
	}

	var (
		buf          []byte = make([]byte, 1024)
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
		h *ipv4.Header
		p []byte
	)

	saddr, sport, dport = c.conf.FriendIP, c.conf.FriendReceivePort, c.conf.OriginReceivePort

	prevPacketTime = time.Now()

	for {

		h, p, _, err = c.readConn(buf)
		if err != nil {
			break
		}
		udph := layers.UDP{}
		if err = udph.DecodeFromBytes(p, gopacket.NilDecodeFeedback); err == nil {
			// We check for the expected source IP, source port, and destination port
			if bytes.Equal(h.Src.To4(), saddr[:]) {
				if udph.SrcPort == layers.UDPPort(sport) && udph.DstPort == layers.UDPPort(dport) {
					if udph.Length == uint16(32) { //end of message
						break
					} else if udph.Length == uint16(34) { //the rest of the message
						var b []byte
						b, err = c.conf.Encoder.GetByte(*h, udph, maskIndex)
						if err != nil {
							break
						}
						if (pos + uint64(len(b))) >= uint64(len(dataBuf)) { //overflow
							err = errors.New("Overflow. End of message never received.")
							break
						} else { //add data to array, increment pos
							for i := 0; i < len(b); i++ {
								dataBuf[pos] = b[i]
								pos++
							}
						}
					} else { //not packet from friend port
						continue
					}
				}
			}
		}
		if c.conf.ReadTimeout > 0 && time.Now().Sub(prevPacketTime) > c.conf.ReadTimeout {
			err = errors.New("Covert Packet Timeout")
			break
		}
	}
	return embedders.CopyData(c.conf.Encoder.GetMask(), pos, dataBuf, data, err)
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
