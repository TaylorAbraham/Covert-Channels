package icmpNormal

import(
	"net"
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
	packetConn net.PacketConn
	clientConn net.Conn
}

// closes the ICMP channel
func (c *Channel) Close() error {
	// close the channel and check if any errors occur
	err := c.packetConn.Close()
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
	packetConn, err := net.ListenPacket("ip4:icmp", (&net.IPAddr{IP: c.conf.OriginIP[:], Port: int(c.conf.OriginPort)}).String())
	if err != nil {
		return nil, err
	}
	c.packetConn = packetConn

	//client connection
	clientConn, err := net.Dial("ip4:icmp", (&net.IPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.DestinationPort)}).String())
	if err != nil {
		return nil, err
	}
	c.clientConn = clientConn

	return c, nil
}

// the server receives the data as it reads 
func (c *Channel) Receive(data []byte) (uint64, error) {
	n, _, err := c.packetConn.ReadFrom(data)
	return uint64(n), err
}

// the client sends the data
func (c *Channel) Send(data []byte) (uint64, error) {
	n, err := c.clientConn.Write(data)
	return uint64(n), err
}
