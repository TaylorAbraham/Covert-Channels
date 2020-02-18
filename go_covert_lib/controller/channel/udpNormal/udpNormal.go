package UdpIP

import(
	"net"
	"strconv"
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
}

// Create channel
func MakeChannel(conf Config) (*Channel, error) {
	var err error

	c := &Channel{
		conf: conf,
	}

	//server ready for incoming udp interaction to server address
	PacketConn, err := net.ListenPacket("udp4", strconv.Itoa(int(c.conf.OriginPort)))
	if err != nil {
		return nil, err
	}

	//client conn
	ClientConn, err := net.Dial("udp4", strconv.Itoa(int(c.conf.DestinationPort)))
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Channel) Receive(data []byte) (uint64, error) {}

func (c *Channel) Send(data []byte) (uint64, error) {}