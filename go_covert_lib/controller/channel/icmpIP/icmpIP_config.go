package icmpIP

import (
	"../../config"
	"errors"
	"time"
)

type ConfigClient struct {
	FriendIP          config.IPV4Param
	OriginIP          config.IPV4Param
	Encoder           config.SelectParam
	WriteTimeout      config.U64Param
	ReadTimeout       config.U64Param
	DialTimeout       config.U64Param
	Identifier	config.U16Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP:          config.MakeIPV4("127.0.0.1", config.Display{Description: "Your friend's IP address.", Name: "Friend's IP", Group: "IP Addresses"}),
		OriginIP:          config.MakeIPV4("127.0.0.1", config.Display{Description: "Your IP address.", Name: "Your IP", Group: "IP Addresses"}),
		WriteTimeout:      config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The write timeout in milliseconds.", Name: "Write Timeout", Group: "Timing"}),
		ReadTimeout:       config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The read timeout in milliseconds.", Name: "Read Timeout", Group: "Timing"}),
		DialTimeout:       config.MakeU64(500, [2]uint64{0, 65535}, config.Display{Description: "The dial timeout for the send method in milliseconds. Zero for no timeout.", Name: "Dial Timeout", Group: "Timing"}),
		Encoder:           config.MakeSelect("id", []string{"id"}, config.Display{Description: "The encoding mechanism to use for this protocol.", Name: "Encoding", Group: "Settings"}),
		Identifier:	config.MakeU16(1234, [2]uint16{0, 65535}, config.Display{Description: "A unique key to distingish covert ICMP packets from other ICMP packets", Name: "Identifier", Group: "Timing"}),
	}
}

func ToChannel(cc ConfigClient) (*Channel, error) {
	var c Config
	var friendIP, originIP [4]byte
	
	c.FriendIP = friendIP
	c.OriginIP = originIP

	c.DialTimeout = time.Duration(cc.DialTimeout.Value) * time.Millisecond
	c.ReadTimeout = time.Duration(cc.ReadTimeout.Value) * time.Millisecond
	c.WriteTimeout = time.Duration(cc.WriteTimeout.Value) * time.Millisecond
	c.Identifier = cc.Identifier.Value

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
