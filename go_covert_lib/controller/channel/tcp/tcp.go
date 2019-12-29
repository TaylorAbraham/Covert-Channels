package tcp

import (
	"bytes"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"math/rand"
	"net"
	"time"
  "strconv"
)

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
	FriendIP   [4]byte
	OriginIP   [4]byte
  FriendReceivePort uint16
  OriginReceivePort uint16
  Encoder     TcpEncoder
  WriteTimeout     time.Duration
  ReadTimeout      time.Duration
}

// A TCP covert channel
type Channel struct {
	conf        Config
	rawConn     *ipv4.RawConn
  listener    net.Listener
  recvChan    chan data
  // A channel for receiving TCP acks from the socket we are sending messages to
  // This allows us to find the Ack and Seq numbers
  replyChan    chan data
	cancel      chan bool
}

type TcpEncoder interface {
  GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error)
  SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, error)
}

// Encoder stores one byte per packet in the lowest order byte of the IPV4 header ID
type IDEncoder struct {}
func (id *IDEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error) {
  return []byte{byte(ipv4h.ID & 0xFF)}, nil
}
func (id *IDEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, error) {
  if len(buf) == 0 {
    return ipv4h, tcph, nil, errors.New("Cannot set byte if no data")
  }
  r := rand.New(rand.NewSource(time.Now().UnixNano()))
  ipv4h.ID = (r.Int() & 0xFF00) | int(buf[0])
  return ipv4h, tcph, buf[1:], nil
}

// Create the covert channel, filling in the SeqEncoder
// with a default if none is provided
// Although the error is not yet used, it is anticipated
// that this function may one day be used for validating
// the data structure
func MakeChannel(conf Config) (*Channel, error) {
	c := &Channel{conf: conf, cancel: make(chan bool), recvChan : make(chan data, 1024), replyChan : make(chan data, 1024)}

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

  //  A loop to read incoming packets and routing them to the appropriate destination
  go c.readLoop()

	return c, nil
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

func (c *Channel) Receive(data []byte, progress chan<- uint64) (uint64, error) {

  for {
    select {
    case <-c.cancel:
      return 0, errors.New("Receive Cancelled")
    default:
    }

    conn, err := c.listener.Accept()
    if err != nil {
      return 0, errors.New("Accept Error")
    }

    var (
//    ack uint32
//    seq uint32
      friendPort uint16
      handshake byte = 0
      n   uint64
      valid bool
      fin   bool
    )

    if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); !ok {
      conn.Close()
      continue
    } else if tcpAddr.IP == nil || bytes.Compare(tcpAddr.IP, c.conf.OriginIP[:]) != 0 {
      conn.Close()
      continue
    } else {
      friendPort = uint16(tcpAddr.Port)
    }

    defer conn.Close()

    // If we have not set a timeout
    if c.conf.ReadTimeout == 0 {
      for {
        select {
        case d := <-c.recvChan:
          n, handshake, valid, fin, err = c.handleReceivedPacket(d, data, n, friendPort, handshake)
          // If there is an error or the fin packet has been received we return
          if err != nil || fin {
            return n, err
          }
        case <-c.cancel:
          return n, errors.New("Receive cancelled")
        }
      }
    } else {
      // Timeouts can occur for two reasons.
      // Either no packet arrives within the timeout period,
      // or a packet arrives but it is not a valid part of the message
      // We cover the first case with the time.After in the select
      // The second is by measuring time an invalid packet arrives since the
      // previous valid packet
      var tValid time.Time = time.Now()
      for {
        select {
        case d := <-c.recvChan:
          n, handshake, valid, fin, err = c.handleReceivedPacket(d, data, n, friendPort, handshake)
          // If there is an error or the fin packet has been received we return
          if err != nil || fin {
            return n, err
          } else if valid {
            tValid = time.Now()
          } else if time.Now().Sub(tValid) > c.conf.ReadTimeout {
            return n, errors.New("Timeout Error")
          }
        case <-time.After(c.conf.ReadTimeout):
          return n, errors.New("Timeout Error")
        case <-c.cancel:
          return n, errors.New("Receive cancelled")
        }
      }
    }
  }
}

func (c *Channel) handleReceivedPacket(d data, data []byte, n uint64, friendPort uint16, handshake byte) (uint64, byte, bool, bool, error) {

  var (
    valid bool  // Was this packet a valid part of the message
    fin bool
    err error
    receivedBytes []byte
  )
  // Verify that the source port is the one associated with this connection
  if layers.TCPPort(friendPort) != d.tcph.SrcPort {
    // Incorrect source Port
  } else if handshake == 0 && d.tcph.SYN {
    // three way handshake SYN
    handshake = 1
    valid = true
  } else if handshake == 1 && d.tcph.ACK {
    // three way handshake ACK
//          ack = d.tcph.Ack
//          seq = d.tcph.Seq
    handshake = 2
  } else if d.tcph.RST {
    // If rst packet is sent we quit
    err = errors.New("RST packet")
  } else if d.tcph.SYN {
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
    // Not started
  } else if d.tcph.FIN {
    valid = true
    // FIN - done message
    fin = true
  } else {
    // Normal transmission packet
    if receivedBytes, err = c.conf.Encoder.GetByte(d.ipv4h, d.tcph); err != nil {
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
    Timeout   : c.conf.WriteTimeout,
  }
  // We must empty the reply chan of all messages
  // That way we will be able to take in new packets
  // when after starting the dial and be able to determine the
  // initial sequence and acknowledgement numbers
  empty:
  for {
    select {
    case <-c.replyChan:
    default:
      break empty
    }
  }

  // DialContext
  conn, err := nd.Dial("tcp4", (&net.TCPAddr{IP : c.conf.FriendIP[:], Port : int(c.conf.FriendReceivePort)}).String())
  if err != nil {
    return 0, err
  }

  defer conn.Close()

  var (
    seq uint32
    ack uint32
    // A slice containing the bytes that have not yet been sent
    rem []byte = data
    n uint64
    originPort uint16
  )

  if tcpAddr, ok := conn.LocalAddr().(*net.TCPAddr); !ok {
    return 0, errors.New("Not TCPAddr")
  } else {
    originPort = uint16(tcpAddr.Port)
  }

  // Wait for the SYN/ACK packet
  wait:
  for {
    select {
    case d := <-c.replyChan:
      // Verify that this packet was sent to the correct address
      if layers.TCPPort(originPort) != d.tcph.DstPort {
        continue
      } else if (d.tcph.ACK && d.tcph.SYN) {
        seq = d.tcph.Ack
        ack = d.tcph.Seq + 1
        break wait
      }
    case <-time.After(time.Second*5):
      return 0, errors.New("ACK packet not received")
    case <-c.cancel:
      return 0, errors.New("Receive cancelled")
    }
  }

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
    SrcIP:    sip[:],
    DstIP:    dip[:],
  }

  tcph.SrcPort = layers.TCPPort(sport)
	tcph.DstPort = layers.TCPPort(dport)

  tcph.Seq = seq
  tcph.Ack = ack

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

func (c *Channel) readLoop() {
  var buf [1024]byte
  for {
    h, p, _, err := c.rawConn.ReadFrom(buf[:])

    select {
    case <-c.cancel:
      close(c.recvChan)
      close(c.replyChan)
      return
    default:
      if err != nil {
        continue
      }
    }

    tcph := layers.TCP{}
    if err = tcph.DecodeFromBytes(p, gopacket.NilDecodeFeedback); err == nil {
      if bytes.Equal(h.Src.To4(), c.conf.OriginIP[:]) {
        var pckChan chan data
        // Packets may be received as ACKs to our sent messages or as packets sent to this channel
        // We must route accordingly
        if tcph.SrcPort == layers.TCPPort(c.conf.FriendReceivePort) {
          // If it is from the port we send to, then it must be an ACK to a packet we sent
          pckChan = c.replyChan
        } else if tcph.DstPort == layers.TCPPort(c.conf.OriginReceivePort) {
          // Otherwise it is potentially from the sender to this channel
          // We confirm by checking the port number based on the port of the open connection
          pckChan = c.recvChan
        } else {
          continue
        }

        select {
          // If our buffered chan is full, empty it
        case pckChan <- data{ipv4h : *h, tcph : tcph}:
        case <-time.After(time.Millisecond*200):
          // If the channel buffer is full we empty the first item and add
          // the current item to the end
          <-pckChan
          pckChan <- data{ipv4h : *h, tcph : tcph}
        }
      }
    }
  }
}

type data struct {
  ipv4h ipv4.Header
  tcph  layers.TCP
}
