package udpIP

import (
	"net"
	"golang.org/x/net/ipv4"
	"strconv"
)

type Config struct {
	FriendIP          [4]byte
	OriginIP          [4]byte
	FriendReceivePort uint16
	OriginReceivePort uint16
}

type Channel struct {
	conf     Config
	rawConn  *ipv4.RawConn
	listener net.Listener
}

//create channel
func MakeChannel(conf Config) (*Channel, error) {

	c := &Channel{conf: conf}

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

	c.listener, err = net.Listen("", ":"+strconv.Itoa(int(c.conf.OriginReceivePort)))
	if err != nil {
		return nil, err
	}

	return c, nil
}