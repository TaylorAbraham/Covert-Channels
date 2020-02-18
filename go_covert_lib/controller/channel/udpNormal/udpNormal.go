package udpNormal

import(
	"net"
)

type Config struct {
	FriendIP          [4]byte
	OriginIP          [4]byte
	DestinationPort uint16
	OriginPort uint16
}

// A UDP channel
type Channel struct {
	conf     Config
	packetConn net.PacketConn
	clientConn net.Conn
}

func (c *Channel) Close() error {
	err := c.packetConn.Close()
	if(err != nil) {
		c.clientConn.Close()
		return err
	}
	err = c.clientConn.Close()

	return err
}

// Create channel
func MakeChannel(conf Config) (*Channel, error) {
	var err error

	c := &Channel{
		conf: conf,

	}

	//server ready for incoming udp interaction to server address
	packetConn, err := net.ListenPacket("udp4", (&net.UDPAddr{IP: c.conf.OriginIP[:], Port: int(c.conf.OriginPort)}).String())
	if err != nil {
		return nil, err
	}
	c.packetConn = packetConn

	//client conn
	clientConn, err := net.Dial("udp4", (&net.UDPAddr{IP: c.conf.FriendIP[:], Port: int(c.conf.DestinationPort)}).String())
	if err != nil {
		return nil, err
	}
	c.clientConn = clientConn

	return c, nil
}

func (c *Channel) Receive(data []byte) (uint64, error) {

	//server reads
	n, _, err := c.packetConn.ReadFrom(data)
	return uint64(n), err
}

func (c *Channel) Send(data []byte) (uint64, error) {
	//client sends
	n, err := c.clientConn.Write(data)
	return uint64(n), err

}