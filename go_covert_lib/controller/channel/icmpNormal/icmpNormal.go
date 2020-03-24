package icmpNormal

import(
	"net"
	"errors"
	"golang.org/x/net/ipv4"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Config struct {
	FriendIP          [4]byte
	OriginIP          [4]byte
	DestinationPort uint16
	OriginPort uint16
}

// This is a normal, non-covert IMCP messaging channel
// The message is sent using normal simiply ICMP packets
type Channel struct {
	conf     Config
	clientConn net.Conn
	rawConn     *ipv4.RawConn
}

// closes the ICMP channel
func (c *Channel) Close() error {
	// close the channel and check if any errors occur
	err := c.rawConn.Close()
	if(err != nil) {
		c.clientConn.Close()
		return err
	}
	err = c.clientConn.Close()

	return err
}

// Create the ICMP normal covert channel
// a default channel that just listens and sends packets
func MakeChannel(conf Config) (*Channel, error) {
	var err error

	c := &Channel{
		conf: conf,
	}

	//server ready for incoming icmp interaction to server address
	packetConn, err := net.ListenPacket("ip4:icmp", (&net.IPAddr{IP: c.conf.OriginIP[:]}).String())
	if err != nil {
		return nil, err
	}

	c.rawConn, err = ipv4.NewRawConn(packetConn)
	if err != nil {
		packetConn.Close()
		return nil, err
	}

	//client connection
	clientConn, err := net.Dial("ip4:icmp", (&net.IPAddr{IP: c.conf.FriendIP[:]}).String())
	if err != nil {
		return nil, err
	}
	c.clientConn = clientConn

	return c, nil
}

// the server receives the data as it reads
func (c *Channel) Receive(data []byte) (uint64, error) {
	buf := make([]byte, 1024)
	_, b, _, err := c.rawConn.ReadFrom(buf)
	if err != nil {
		return 0, err
	}
	b = b[8:]
	copy(data, b)
	if len(b) > len(data) {
		return uint64(len(data)), errors.New("Buffer Overflow")
	} else {
		return uint64(len(b)), nil
	}
}

// the client sends the data
func (c *Channel) Send(data []byte) (uint64, error) {

	var icmph layers.ICMPv4 = layers.ICMPv4{
		TypeCode : layers.CreateICMPv4TypeCode(1, 0),
	}

	sb := gopacket.NewSerializeBuffer()
	op := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	if err := gopacket.SerializeLayers(sb, op, &icmph, gopacket.Payload(data)); err != nil {
		return 0, err
	}

	n, err := c.clientConn.Write(sb.Bytes())

	if n >= 8 {
		n = n - 8
	} else {
		n = 0
	}

	return uint64(n), err
}
