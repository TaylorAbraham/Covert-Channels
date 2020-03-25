package icmpNormal

import (
	"../../config"
	"errors"
)

type ConfigClient struct {
	FriendIP        config.IPV4Param
	OriginIP        config.IPV4Param
	DestinationPort config.U16Param
	OriginPort      config.U16Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP:        config.MakeIPV4("127.0.0.1", config.Display{Description: "Your destination IP Address."}),
		OriginIP:        config.MakeIPV4("127.0.0.1", config.Display{Description: "Your IP Address."}),
		DestinationPort: config.MakeU16(8123, [2]uint16{0, 65535}, config.Display{Description: "Your friends ICMP receive Port. Their send port is chosen randomly."}),
		OriginPort:      config.MakeU16(8124, [2]uint16{0, 65535}, config.Display{Description: "Your ICMP receive Port. Send port is chosen randomly."}),
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
	c.DestinationPort = cc.DestinationPort.Value
	c.OriginPort = cc.OriginPort.Value

	if ch, err := MakeChannel(c); err != nil {
		return nil, err
	} else {
		return ch, nil
	}
}
