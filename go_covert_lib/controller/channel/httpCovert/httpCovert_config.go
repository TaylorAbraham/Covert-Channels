package httpCovert

import (
	"errors"
	"time"

	"../../config"
)

type ConfigClient struct {
	FriendIP       config.IPV4Param
	OriginIP       config.IPV4Param
	FriendPort     config.U16Param
	OriginPort     config.U16Param
	UserType       config.SelectParam
	ClientPollRate config.U64Param
	ClientTimeout  config.U64Param
	ReadTimeout    config.U64Param
	WriteTimeout   config.U64Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP:       config.MakeIPV4("127.0.0.1", config.Display{Description: "Your friend's IP address.", Name: "Friend's IP", Group: "IP Addresses"}),
		OriginIP:       config.MakeIPV4("127.0.0.1", config.Display{Description: "Your IP address.", Name: "Your IP", Group: "IP Addresses"}),
		FriendPort:     config.MakeU16(8123, [2]uint16{0, 65535}, config.Display{Description: "Your friend's port.", Name: "Friend's Port", Group: "Ports"}),
		OriginPort:     config.MakeU16(8124, [2]uint16{0, 65535}, config.Display{Description: "Your port.", Name: "Your Port", Group: "Ports"}),
		UserType:       config.MakeSelect("client", []string{"client", "server"}, config.Display{Description: "Should this covert channel act as a client or server?", Name: "UserType", Group: "Settings"}),
		ClientPollRate: config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The poll rate in milliseconds.", Name: "Poll Rate", Group: "Timing"}),
		ClientTimeout:  config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The client timeout in milliseconds.", Name: "Client Timeout", Group: "Timing"}),
		WriteTimeout:   config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The write timeout in milliseconds.", Name: "Write Timeout", Group: "Timing"}),
		ReadTimeout:    config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The read timeout in milliseconds.", Name: "Read Timeout", Group: "Timing"}),
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
	c.FriendPort = cc.FriendPort.Value
	c.OriginPort = cc.OriginPort.Value
	c.ClientPollRate = time.Duration(cc.ClientPollRate.Value) * time.Millisecond
	c.ClientTimeout = time.Duration(cc.ClientTimeout.Value) * time.Millisecond
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
