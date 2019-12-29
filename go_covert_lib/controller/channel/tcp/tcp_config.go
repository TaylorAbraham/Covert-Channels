package tcp

import (
	"../../config"
	"errors"
	"time"
)

type ConfigClient struct {
	FriendIP     config.IPV4Param
	OriginIP     config.IPV4Param
	FriendReceivePort   config.U16Param
	OriginReceivePort   config.U16Param
	Encoder      config.SelectParam
	WriteTimeout config.U64Param
	ReadTimeout  config.U64Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP:     config.MakeIPV4("127.0.0.1", config.Display{Description: "Your friends IP Address."}),
		OriginIP:     config.MakeIPV4("127.0.0.1", config.Display{Description: "Your IP Address."}),
		FriendReceivePort:   config.MakeU16(8123, [2]uint16{0, 65535}, config.Display{Description: "Your friends tcp receive Port. Their send port is chosen randomly."}),
		OriginReceivePort:   config.MakeU16(8124, [2]uint16{0, 65535}, config.Display{Description: "Your tcp receive Port. Send port is chosen randomly."}),
		Encoder:      config.MakeSelect("id", []string{"id"}, config.Display{Description: "The encoding mechanism to use for this protocol."}),
		WriteTimeout: config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The write timeout in milliseconds."}),
		ReadTimeout:  config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The read timeout in milliseconds."}),
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

	c.WriteTimeout = time.Duration(cc.WriteTimeout.Value)
	c.ReadTimeout = time.Duration(cc.ReadTimeout.Value)

	switch cc.Encoder.Value {
	case "id":
		c.Encoder = &IDEncoder{}
	default:
		return nil, errors.New("Invalid encoder value")
	}

	if ch, err := MakeChannel(c); err != nil {
		return nil, err
	} else {
		return ch, nil
	}
}
