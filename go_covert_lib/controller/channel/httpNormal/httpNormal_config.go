package httpNormal

import (
	"errors"
	"time"

	"../../config"
)

type ConfigClient struct {
	FriendIP          config.IPV4Param
	OriginIP          config.IPV4Param
	FriendReceivePort config.U16Param
	OriginReceivePort config.U16Param
	UserType          config.SelectParam
	ReadTimeout       config.U64Param
	WriteTimeout      config.U64Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP:          config.MakeIPV4("127.0.0.1", config.Display{Description: "Your friends IP Address."}),
		OriginIP:          config.MakeIPV4("127.0.0.1", config.Display{Description: "Your IP Address."}),
		FriendReceivePort: config.MakeU16(8123, [2]uint16{0, 65535}, config.Display{Description: "Your friends tcp receive Port. Their send port is chosen randomly."}),
		OriginReceivePort: config.MakeU16(8124, [2]uint16{0, 65535}, config.Display{Description: "Your tcp receive Port. Send port is chosen randomly."}),
		UserType:          config.MakeSelect("client", []string{"client", "server"}, config.Display{Description: "Should this covert channel act as a client or server?", Name: "UserType", Group: "Settings"}),
		ReadTimeout:       config.MakeU64(500, [2]uint64{0, 65535}, config.Display{Description: "The intra-packet read timeout for the receive method in milliseconds. Zero for no timeout."}),
		WriteTimeout:      config.MakeU64(500, [2]uint64{0, 65535}, config.Display{Description: "The a timeout for writing packets to the raw socket, in milliseconds. Zero for no timeout."}),
	}
}

func ToChannel(cc ConfigClient) (*Channel, error) {
	var c Config
	var friendIP, originIP [4]byte
	var err error
	if friendIP, err = cc.FriendIP.GetValue(); err != nil {
		return nil, errors.New("Invalid FriendIP value")
	}
	if originIP, err = cc.OriginIP.GetValue(); err != nil {
		return nil, errors.New("Invalid OriginIP value")
	}

	c.FriendIP = friendIP
	c.OriginIP = originIP
	c.FriendReceivePort = cc.FriendReceivePort.Value
	c.OriginReceivePort = cc.OriginReceivePort.Value
	c.ReadTimeout = time.Duration(cc.ReadTimeout.Value) * time.Millisecond
	c.WriteTimeout = time.Duration(cc.WriteTimeout.Value) * time.Millisecond

	switch cc.UserType.Value {
	case "client":
		c.UserType = 0
	case "server":
		c.UserType = 1
	default:
		return nil, errors.New("Invalid user type value")
	}

	if ch, err := MakeChannel(c); err != nil {
		return nil, err
	} else {
		return ch, nil
	}
}
