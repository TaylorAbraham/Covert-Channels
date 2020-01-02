package tcp

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const maxAccept = 32

// Based on experimental results,
// a raw socket can hold approximately 300 packets
// before they start to get dropped.
// This number was chosen based on that estimate of how
// many packets should be stored
const maxStorePacket = 512

type packet struct {
	ipv4h ipv4.Header
	tcph  layers.TCP
}

type acceptedConn struct {
	conn net.Conn
	// The TCP port used by our Friend IP in this covert message
	friendPort uint16
}

type portRequest struct {
	// The friendPort retrieved from the open TCP connection
	// in the Receive method
	port uint16
	// A go channel to allow the portRouter
	// to provide the desired packet go channel to the Receive method
	reply chan chan packet
}

type dropRequest struct {
	port     uint16
	keepNext bool
}

// This covert channel is formed by two peers.
// Each opens a tcp socket and listens for incomming connections when
// when receiving
// To send a message to the other peer in the covert channel, this covert
// channel dials a tcp connection to the others listening socket.
// Once dialing is done, a raw socket is used to send tcp packets along
// the packets have the correct seq and ack numbers (found by observing the
// three-way handshake) and have data hidden within the fields of the IP or TCP
// header. When done sending, the sender side closes their TCP connection, causing
// the four way close handshake. The receiver reads the FIN packet and exits.
// The port numbers of the listening TCP sockets must be specified to open
// valid TCP sockets. However, the TCP port of the dialer is not specified, and
// is chosen (usually pseudorandomly) by the dialer. During the three way handshake
// the src port is read by the receiver and used to filter all subsequent TCP packets.
// Using a randomly generated TCP dial port seems to be necessary for proper use.
// When a TCP client is closed, there seems to be a period of time when the same
// port cannot be reused. If a sender port were assigned, then calls
// to the Send method would fail if performed in quick succession, reducing the
// rate that messages could be sent. Using a random sender port completely
// alleviates this risk.
type Config struct {
	FriendIP          [4]byte
	OriginIP          [4]byte
	FriendReceivePort uint16
	OriginReceivePort uint16
	Encoder           TcpEncoder
	// The TCP dial timeout for the three way handshake in the the send method
	DialTimeout time.Duration
	// The intra-packet read timeout.
	// The receive method will block until a three way handshake
	// is complete and the listener returns, and will only exit with a
	// read timeout if the intra-packet delay is too large.
	ReadTimeout time.Duration
}

// A TCP covert channel
type Channel struct {
	conf     Config
	rawConn  *ipv4.RawConn
	listener net.Listener

	// A channel to close the covert channel
	// This must be the only go channel that is closed
	// That way only a single signal can be used for signalling
	// the covert channel close operation.
	// Otherwise errors could occur with invalid data being
	// received on the other go channels
	// Although there are workarounds, it is simplest to just
	// guarantee that the Close() signal is only mediated by this channel
	cancel chan bool

	// A go channel for receiving TCP acks from the socket we are sending messages to
	// This allows us to find the Ack and Seq numbers
	replyChan     chan packet
	receiveRouter portRouter
	sendRouter    portRouter

	acceptChan chan acceptedConn
}

type TcpEncoder interface {
	GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error)
	SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, error)
}

// Encoder stores one byte per packet in the lowest order byte of the IPV4 header ID
type IDEncoder struct{}

func (id *IDEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error) {
	return []byte{byte(ipv4h.ID & 0xFF)}, nil
}
func (id *IDEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, errors.New("Cannot set byte if no data")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ipv4h.ID = (r.Int() & 0xFF00) | int(buf[0])
	// Based on my experimental results, the raw socket will override
	// an IP ID of zero. We use this loop to ensure that the ID is something
	// other than zero so that our real data is transmitted
	for ipv4h.ID == 0 {
		ipv4h.ID = (r.Int() & 0xFF00) | int(buf[0])
	}

	return ipv4h, tcph, buf[1:], nil
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	c := &Channel{conf: conf,
		cancel: make(chan bool),

		// Only 32 connections can be accepted before they begin to be dropped
		acceptChan: make(chan acceptedConn, maxAccept),

		receiveRouter: portRouter{
			pktRecvChan:     make(chan packet, 1024),
			requestPortChan: make(chan portRequest),
			dropPortChan:    make(chan dropRequest),
		},
		sendRouter: portRouter{
			pktRecvChan:     make(chan packet, 1024),
			requestPortChan: make(chan portRequest),
			dropPortChan:    make(chan dropRequest),
		},
	}

	if c.conf.Encoder == nil {
		c.conf.Encoder = &IDEncoder{}
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

	c.listener, err = net.Listen("tcp4", ":"+strconv.Itoa(int(c.conf.OriginReceivePort)))
	if err != nil {
		return nil, err
	}

	// A loop to read incoming packets and routing them to the appropriate destination
	go c.readLoop(c.receiveRouter.pktRecvChan, c.sendRouter.pktRecvChan)
	// A loop to accept incoming tcp connection
	go c.acceptLoop()
	// A loop to separate out incoming packets for the Receive method (i.e. not the replies to the Send method)
	// based on their src port
	go c.receiveRouter.run(true, c.cancel)

	go c.sendRouter.run(false, c.cancel)

	return c, nil
}

type portRouter struct {
	// pktRecvChan must be a buffered channel to
	// allow the readLoop to keep reading even while the router processes
	// sorting the packets by port
	pktRecvChan chan packet
	// Allows a caller to request a go channel for receiving from the
	// desired friend port
	requestPortChan chan portRequest
	// A channel to signal to the portRouter loop that
	// a given message has been completely received from the given friend port,
	// and that the storage space for that friend port may be dropped
	dropPortChan chan dropRequest
}

func (c *Channel) Close() error {
	select {
	// Have we already closed
	case <-c.cancel:
		return nil
	default:
		close(c.cancel)
	}
	if err := c.rawConn.Close(); err != nil {
		c.listener.Close()
		return err
	}
	return c.listener.Close()
}

// A convenience function to handle the two timeout scenarios.
// Incoming packets are handled by a separate go routine.
// When a new packet arrives it is sent on a channel, the receiver then
// validates it. Thus, a timeout could occur if no packet is received over the
// go channel in enough time or if there are packets coming along the go channel
// but they are not valid packets. We check for both cases here.
// Set timeout to zero for no timeout
func waitPacket(pktChan chan packet, timeout time.Duration, f func(p packet) (bool, error), cancel chan bool) error {
	var startTime time.Time = time.Now()
	if timeout == 0 {
		for {
			select {
			case p := <-pktChan:
				ok, err := f(p)
				// Exit if done (i.e. a valid packet is received and processed) or an error is received
				if ok || err != nil {
					return err
				}
			case <-cancel:
				return errors.New("Cancel")
			}
		}
	} else {
		for {
			select {
			case p := <-pktChan:
				ok, err := f(p)
				// Exit if done (i.e. a valid packet is received and processed) or an error is received
				if ok || err != nil {
					return err
				} else if time.Now().Sub(startTime) > timeout {
					return errors.New("Timeout")
				}
			case <-time.After(timeout):
				return errors.New("Timeout")
			case <-cancel:
				return errors.New("Cancel")
			}
		}
	}
}

func (c *Channel) Receive(data []byte, progress chan<- uint64) (uint64, error) {

	var (
		ac acceptedConn
		//    ack uint32
		//    seq uint32
		// When we read packets with the raw socket we read the packets of the 3-way handshake
		// This variable is used to keep track of the stage in the 3-way handshake
		handshake byte   = 0
		n         uint64 // the number of bytes received
		fin       bool   // if the FIN packet has arrived
	)

	// Check if we should timeout
	if c.conf.DialTimeout == 0 {
		select {
		case ac = <-c.acceptChan:
		// Check if the covert channel has been closed
		case <-c.cancel:
			return 0, errors.New("Receive Cancelled")
		}
	} else {
		select {
		case ac = <-c.acceptChan:
		// Check if the covert channel has been closed
		case <-c.cancel:
			return 0, errors.New("Receive Cancelled")
		case <-time.After(c.conf.DialTimeout):
			return 0, errors.New("Accept timeout")
		}
	}

	defer ac.conn.Close()

	// Once a TCP connection is accepted,
	// we must retrieve a go channel from the portRouter loop
	// that only provides packets from the desired friendPort
	recvPktChan, err := c.receiveRouter.getPktChan(ac.friendPort, c.cancel)
	if err != nil {
		return 0, err
	}

	// Once we are done receiving the message we let the portRouter loop
	// know that it no longer needs to store space for this friend port
	// This will be called before the ac.conn.Close() defer above
	// due to go defer execution order
	// Doing so simplifies the code, but (based on experimental results)
	// increases the risk that trailing packets could arrive
	// after resources for packets from this friendport
	// are dropped
	// This issue is discussed and handled further in the portRouter loop
	defer c.receiveRouter.donePktChan(ac.friendPort, true, c.cancel)

	// Exit when the fin packet is received
	for !fin {
		var err error
		// Wait until a valid packet is received
		// This way we measure the time between valid packets.
		// If no valid packet is received within timeout of the previous valid packet
		// we exit with an error
		err = waitPacket(recvPktChan, c.conf.ReadTimeout, func(p packet) (bool, error) {
			var (
				valid bool
				err   error
			)
			n, handshake, valid, fin, err = c.handleReceivedPacket(p, data, n, ac.friendPort, handshake)
			return valid, err
		}, c.cancel)
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// Once the three way handshake is complete we must read the incoming packets
// Since any packet could arrive, we must distinguish incomming packets
// that are part of the three way handshake and covert message from those that are not.
// The handshake variable allows use to keep track of the three way handshake packets.
// Once it reaches 2 the three way handshake is done and the rest of the packets (from the correct port)
// form the message. A packet with the FIN flag set indicates the end of transmission (due to the closed
// TCP connection). RST or second SYN packets are interpreted as an error in the connection and
// cause the Receive method to abort.
// We return a valid flag to indicate if the packet forms part of the TCP covert communication (three way handshake, message, or FIN packet )
func (c *Channel) handleReceivedPacket(p packet, data []byte, n uint64, friendPort uint16, handshake byte) (uint64, byte, bool, bool, error) {

	var (
		valid         bool // Was this packet a valid part of the message
		fin           bool
		err           error
		receivedBytes []byte
	)
	// Verify that the source port is the one associated with this connection
	if layers.TCPPort(friendPort) != p.tcph.SrcPort {
		// Incorrect source Port
	} else if handshake == 0 && p.tcph.SYN {
		// three way handshake SYN
		handshake = 1
		valid = true
	} else if handshake == 1 && p.tcph.ACK {
		// three way handshake ACK
		//          ack = d.tcph.Ack
		//          seq = d.tcph.Seq
		handshake = 2
	} else if p.tcph.RST {
		// If rst packet is sent we quit
		err = errors.New("RST packet")
	} else if p.tcph.SYN {
		// During testing I found that a hung receiver connection
		// could still be dialed a second time from the sender
		// I.e. the tcp dial operation could be performed in the sender method
		// even though the listener had already opened a connection with the same address
		// In this scenario a FIN packet was likely lost from a previous message.
		// To cope with this scenario, we just abort the current reception if
		// we receive a second SYN packet, loosing preceeding and current message,
		// but resetting the receive call to ensure that subsequent messages can be read
		err = errors.New("Duplicate SYN packet")
	} else if handshake != 2 {
		// Covert message not started since we have not yet reached the packets
		// making up the three way handshake
	} else if p.tcph.FIN {
		valid = true
		// FIN - done message
		fin = true
	} else {
		// Normal transmission packet
		if receivedBytes, err = c.conf.Encoder.GetByte(p.ipv4h, p.tcph); err != nil {
			err = errors.New("Decode Error")
		} else {
			valid = true
			for _, b := range receivedBytes {
				if n < uint64(len(data)) {
					data[n] = b
					n += 1
				} else {
					err = errors.New("Buffer Full")
					break
				}
			}
		}
	}
	return n, handshake, valid, fin, err
}

func (c *Channel) Send(data []byte, progress chan<- uint64) (uint64, error) {

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
	conn, err := nd.DialContext(ctx, "tcp4", (&net.TCPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.FriendReceivePort)}).String())
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	var (
		seq uint32
		ack uint32
		// A slice containing the bytes that have not yet been sent
		rem        []byte = data
		n          uint64
		originPort uint16
	)

	if tcpAddr, ok := conn.LocalAddr().(*net.TCPAddr); !ok {
		return 0, errors.New("Not TCPAddr")
	} else {
		originPort = uint16(tcpAddr.Port)
	}

	recvPktChan, err := c.sendRouter.getPktChan(originPort, c.cancel)
	if err != nil {
		return 0, err
	}
	defer c.sendRouter.donePktChan(originPort, false, c.cancel)

	// Wait for the SYN/ACK packet
	err = waitPacket(recvPktChan, time.Second*5, func(p packet) (bool, error) {
		// We empty packets from the channel until we get the SYN/ACK packet
		// from the 3-way handshake. This packet can be used to retrieve
		// the seq and ack numbers
		if layers.TCPPort(originPort) == p.tcph.DstPort && p.tcph.ACK && p.tcph.SYN {
			seq = p.tcph.Ack
			ack = p.tcph.Seq + 1
			return true, nil
		}
		return false, nil
	}, c.cancel)
	if err != nil {
		return 0, err
	}

	// Send each packet
	for len(rem) > 0 {
		var (
			ipv4h ipv4.Header         = createIPHeader(c.conf.OriginIP, c.conf.FriendIP)
			cm    ipv4.ControlMessage = createCM(c.conf.OriginIP, c.conf.FriendIP)
			tcph  layers.TCP
			wbuf  []byte
		)
		if ipv4h, tcph, rem, err = c.conf.Encoder.SetByte(ipv4h, tcph, rem); err != nil {
			return n, err
		}
		if wbuf, err = createTCPHeader(tcph, seq, ack, c.conf.OriginIP, c.conf.FriendIP, originPort, c.conf.FriendReceivePort); err != nil {
			return n, err
		}
		if err = c.rawConn.WriteTo(&ipv4h, wbuf, &cm); err != nil {
			return n, err
		}
		n = uint64(len(data) - len(rem))
	}
	return n, err
}

func createTCPHeader(tcph layers.TCP, seq, ack uint32, sip, dip [4]byte, sport, dport uint16) ([]byte, error) {

	iph := layers.IPv4{
		SrcIP: sip[:],
		DstIP: dip[:],
	}

	tcph.SrcPort = layers.TCPPort(sport)
	tcph.DstPort = layers.TCPPort(dport)

	tcph.Seq = seq
	tcph.Ack = ack

	// These packets must have the ACK flag set as they are a part of regular messages
	tcph.ACK = true

	// Based on a preliminary investigation of my machine (running Ubuntu 18.04),
	// SYN packets always seem to have a window of 65495
	tcph.Window = 65495

	if err := tcph.SetNetworkLayerForChecksum(&iph); err != nil {
		return nil, err
	}

	sb := gopacket.NewSerializeBuffer()
	op := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	// This will compute proper checksums
	if err := tcph.SerializeTo(sb, op); err != nil {
		return nil, err
	}

	return sb.Bytes(), nil
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

// A loop that continuously receives packets across the raw socket
// Incoming packets are analysed to confirm that they
// have the expected source IP address (Our friend's IP address)
// Based on the port numbers, we
// The receiver chan is for packets destined to calls to the Receive method
// The sender chan is for packets destined to calls to the Send method
func (c *Channel) readLoop(recieverChan chan packet, senderChan chan packet) {
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
			if bytes.Equal(h.Src.To4(), c.conf.FriendIP[:]) {
				var pckChan chan packet
				// Packets may be received as ACKs to our sent messages or as packets sent to this channel
				// We must route accordingly
				if tcph.DstPort == layers.TCPPort(c.conf.OriginReceivePort) {
					// If the port number is to our receive port, then it is potentially
					// an incomming message, so we route it to the Receive methods
					pckChan = recieverChan
				} else if tcph.SrcPort == layers.TCPPort(c.conf.FriendReceivePort) {
					// Otherwise, if the packet is from our friend's receive port, then it is likely an ACK to a packet we sent
					// and we route it to the Send methods
					pckChan = senderChan
				} else {
					continue
				}

				select {
				case pckChan <- packet{ipv4h: *h, tcph: tcph}:
				case <-time.After(time.Millisecond * 200):
					l.Println("Too many packets: Dropped Packet")
				}
			}
		}
	}
}

// A loop to accept incoming TCP connections,
// verify if they are to the correct friend IP address,
// and if so, extract the friend port and send it to the receive method
func (c *Channel) acceptLoop() {
	var l *log.Logger = log.New(os.Stderr, "", log.Flags())
	for {
		select {
		case <-c.cancel:
			for {
				// Ensure all of the connections are closed properly before exiting
				// by emptying the acceptChan and manually closing all TCP connections
				select {
				case ac := <-c.acceptChan:
					ac.conn.Close()
				default:
					return
				}
			}
		default:
		}

		// Accept an incoming TCP connection
		// This allows this covert channel to make use of a proper 3-way TCP handshake,
		// to better obscure the covert communication that is occurring.
		conn, err := c.listener.Accept()
		if err != nil {
			continue
		}

		// Check that the TCP connection is being established from the correct
		// peer IP address
		// If not, we close the connection and continue the receive loop
		if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); !ok {
			conn.Close()
		} else if tcpAddr.IP == nil || bytes.Compare(tcpAddr.IP, c.conf.OriginIP[:]) != 0 {
			conn.Close()
		} else {
			// We have the correct IP address.
			// We must now record the port associated with that address to
			// restrict incoming packets to that port, and not a different port on the same machine.
			select {
			case c.acceptChan <- acceptedConn{conn: conn, friendPort: uint16(tcpAddr.Port)}:
			default:
				l.Println("Too many connections: Dropped TCP Connection")
				conn.Close()
			}
		}
	}
}

type pktStore struct {
	hasRead bool
	pktChan chan packet
}

// Messages could be received from multiple friend ports in quick succession
// if the send method is called repeatedly.
// This method takes the incoming packets from the recvChan
// and groups them based on their source port (the friend port)
// That way we can hold those different messages separately
// and handle them one at a time.
// The Receive method sends along the requestPortChan go channels
// to retrieve a go channel that will provide packets received from
// the desired friend port
// If useSrcPort is true then packets are routed based on their source port
// Otherwise they are routed based on their destination port
func (r *portRouter) run(useSrcPort bool, cancel chan bool) {
	var (
		pktMap map[uint16]*pktStore = make(map[uint16]*pktStore)
		l      *log.Logger          = log.New(os.Stderr, "", log.Flags())
	)

loop:
	for {
		select {
		// The Receive method requests a channel to report
		// packets coming from the desired friend port
		case request := <-r.requestPortChan:
			if _, ok := pktMap[request.port]; ok {
				pktMap[request.port].hasRead = true
				request.reply <- pktMap[request.port].pktChan
			} else if len(pktMap) < maxAccept || dropUnused(pktMap) {
				pktMap[request.port] = &pktStore{
					hasRead: true,
					pktChan: make(chan packet, maxStorePacket),
				}
				request.reply <- pktMap[request.port].pktChan
			} else {
				// else packetChan is nil
				request.reply <- nil
			}
		case p := <-r.pktRecvChan:
			var friendPort uint16
			if useSrcPort {
				friendPort = uint16(p.tcph.SrcPort)
			} else {
				friendPort = uint16(p.tcph.DstPort)
			}
			if _, ok := pktMap[friendPort]; ok {
				// we have already stored packets for this friend port
				// or are waiting for packets on this friend port (by
				// creating the channel with the requestPortChan in the Receive method)
				// We will send the packet on this go channel
			} else if p.tcph.SYN {

				// If a go channel has not already been setup for this friend port,
				// we check if the initial packet is a SYN packet
				// All TCP messages should start with the SYN packet of a three way handshake,
				// so we restrict creating the go channel until the SYN packet is received
				// This helps us protect against any trailing packets from the prceeding message that
				// arrive after the Receive method has terminated and the go channel has been removed
				// by the dropPortChan below
				// It helps prevent us from having trailing resources.

				// If we have room to store this packet go channel or we can clear space for the channel
				if len(pktMap) < maxAccept || dropUnused(pktMap) {
					pktMap[friendPort] = &pktStore{
						hasRead: false,
						pktChan: make(chan packet, maxStorePacket),
					}
				} else {
					// There is not room to store the packet,
					// since all of he go channels are actively
					// being read by a Receive method
					l.Println("Too many connections: Dropped Packet")
					continue
				}
			} else {
				// If a SYN packet is not the first packet in the stream, we drop it silently
				// Such a situation is very common based on experimental results,
				// so it is not worth reporting
				continue
			}

			// Send the packet on the channel
			select {
			case pktMap[friendPort].pktChan <- p:
			case <-time.After(time.Millisecond * 200):
				l.Println("Too many packets: Dropped Packet")
			}
		case drop := <-r.dropPortChan:

			// Under normal operation it should not be possible to
			// attempt to drop a port that has already been dropped,
			// since the request of a port and dropping always occur together
			// in a single Receive or Send method.
			// Depending on how TCP works exactly, it is conceivable that
			// this could arrise for the Receive method. If two subsequent
			// dials are performed from the same friend port (which should theoretically
			// be prohibited by the host machine, but could be circumvented by a crafty
			// attacket), it is possible that
			// the same port and IP numbers for a TCP connection could be accepted twice
			// (this possibility depends on how this machine runs its TCP listener),
			// leading to two Receive methods running for the same friend port
			// Since I am not 100 % sure that such a scenario is impossible, we check
			// here to prevent nil pointer errors below
			if _, ok := pktMap[drop.port]; !ok {
				continue
			}

			// if the keepNext flag is false, we just drop the full channel
			// Otherwise we will employ the algorithm shown below
			// to keep any packets after the first stored SYN packet
			// so as to potentially hold on to a second message that is arriving
			if !drop.keepNext {
				delete(pktMap, drop.port)
				continue
			}

			// Once the Receive method is done, it no longer
			// needs to receive packets from the given friend port
			// and the go channel for receiving such packets may be
			// dropped to make room for other incoming messages from
			// different friend ports

			// First, we check that there are no syn packets in the queue for this
			// friend port.
			// Although the sender should not generally send messages using the same
			// port multiple times in quick succession, it is theoretically possible
			// to do so, and we want to avoid dropping the second message if possible
			// Empty the go channel of all packets, checking for SYN packets

			var (
				hasSYN   bool
				newStore pktStore = pktStore{
					hasRead: false,
					pktChan: make(chan packet, maxStorePacket),
				}
			)
		dropLoop:
			for {
				select {
				case p := <-pktMap[drop.port].pktChan:
					if p.tcph.SYN {
						hasSYN = true
					}
					if hasSYN {
						newStore.pktChan <- p
					}
				default:
					break dropLoop
				}
			}
			// If a SYN packet was found, replace the empty channel with
			// the new channel that has stored all packets from the SYN packet onward
			if hasSYN {
				pktMap[drop.port] = &newStore
			} else {
				delete(pktMap, drop.port)
			}
		case <-cancel:
			break loop
		}
	}
}

func dropUnused(pktMap map[uint16]*pktStore) bool {

	// Everytime a new packet arrives it is added to the pktMap
	// in portRouter so that it may be read by the Receive method
	// when appropriate.
	// This raises a problem, as packet that are not part of a valid
	// message may arrive but be stored in the pktMap
	// If no successful dial (i.e. three way handshake) is associated
	// with the friendPort of those packets, those packets may never be read by a Receive
	// since it will not receive a corresponding connection from the acceptLoop
	// To handle this situation, this function is available to remove unused packets
	// if a new packet (with a new friendPort) arrives and there is not enough room in

	var storedLen map[uint16]int = make(map[uint16]int)

	var minPort uint16
	var minLen int

	for storedPort := range pktMap {
		if !pktMap[storedPort].hasRead {
			storedLen[storedPort] = len(pktMap[storedPort].pktChan)
			// We assign the minLen to guarantee that the below for
			// loop always includes a valid port key
			// This length will be used as a baseline for calculating
			// the actual minimum length
			minLen = storedLen[storedPort]
		}
	}

	if 0 == len(storedLen) {
		// If all all of the pktStores have been read from, we must drop the incoming packet because
		// the go channel is part of an ongoing connection
		// If not, we are permitted to drop one of the packet stores to make room.
		return false
	} else {
		// we drop the pktStore with he fewest stored packets in the go channel,
		// as that is more likely to represent an incomplete message
		for k, v := range storedLen {
			if v <= minLen {
				minLen = v
				minPort = k
			}
		}
		delete(pktMap, minPort)
		return true
	}
}

func (r *portRouter) getPktChan(port uint16, cancel chan bool) (chan packet, error) {
	// This code requests the specific friendPort
	// by providing a go channel to allow portRouter to
	// reply with the go packet channel
	var reply chan chan packet = make(chan chan packet)
	select {
	case r.requestPortChan <- portRequest{
		port:  port,
		reply: reply,
	}:
	case <-cancel:
		return nil, errors.New("Receive Cancelled")
	}

	var recvPktChan chan packet
	// If the above select successfully sent a request on requestPortChan, then we will
	// then we will always get a reply on the reply go chan
	// There is thus no need to select on the cancel go channel
	recvPktChan = <-reply
	// If nil is returned, it means there was not sufficient space
	// to store a go channel for this friend port
	if recvPktChan == nil {
		return nil, errors.New("Insufficent space to receive packets from this friend port")
	} else {
		return recvPktChan, nil
	}
}

// Function to drop the resources accosiated with storing packets from a given port
func (r *portRouter) donePktChan(port uint16, keepNext bool, cancel chan bool) {
	select {
	case r.dropPortChan <- dropRequest{port: port, keepNext: keepNext}:
	case <-cancel:
	}
}
