package ipv4TCP

import (
	"errors"
	"time"
	"../../config"
)

type ConfigClient struct {
	FriendIP   config.IPV4Param
	OriginIP   config.IPV4Param
	BounceIP   config.IPV4Param
	FriendPort config.U16Param
	OriginPort config.U16Param
	BouncePort config.U16Param
	Bounce 		 config.BoolParam
	Delimiter  config.SelectParam
	Encoder    config.SelectParam
	GetDelay   config.SelectParam
	WriteTimeout config.U64Param
	ReadTimeout  config.U64Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP   : config.MakeIPV4("127.0.0.1", "Your Friends IP Address"),
		OriginIP   : config.MakeIPV4("127.0.0.1", "Your IP Address"),
		BounceIP   : config.MakeIPV4("127.0.0.1", "The Bouncer IP Address, if any"),
		FriendPort : config.MakeU16(8123, [2]uint16{0, 65535}, "Your Friends Port"),
		OriginPort : config.MakeU16(8124, [2]uint16{0, 65535}, "Your Port"),
		BouncePort : config.MakeU16(0, [2]uint16{0, 65535}, "The Bouncer Port, if any"),
		Bounce 		 : config.MakeBool(false, "Whether or not we are in bounce mode"),
		Delimiter  : config.MakeSelect("protocol", []string{"buffer", "protocol"}, "The delimiter to use for deciding when to return after having received a message"),
		Encoder    : config.MakeSelect("sequence", []string{"sequence"}, "The encoding mechanism to use for this protocol"),
		GetDelay   : config.MakeSelect("none", []string{"none"}, "The function to use for inter byte delay"),
		WriteTimeout : config.MakeU64(0, [2]uint64{0, 65535}, "The Write Timeout in milliseconds"),
		ReadTimeout  : config.MakeU64(0, [2]uint64{0, 65535}, "The Read Timeout in milliseconds"),
	}
}

func ToChannel (cc ConfigClient) (*Channel, error) {
	var c Config
	var friendIP, originIP, bounceIP [4]byte
	var err error
	if friendIP, err = cc.FriendIP.GetValue(); err != nil {
		return nil, errors.New("Invalid FriendIP value")
	}
	if originIP, err = cc.OriginIP.GetValue(); err != nil {
		return nil, errors.New("Invalid OriginIP value")
	}
	if bounceIP, err = cc.BounceIP.GetValue(); err != nil {
		return nil, errors.New("Invalid BounceIP value")
	}
	c.FriendIP = friendIP
	c.OriginIP = originIP
	c.BounceIP = bounceIP
	c.FriendPort = cc.FriendPort.Value
	c.OriginPort = cc.OriginPort.Value
	c.BouncePort = cc.BouncePort.Value
	c.Bounce = cc.Bounce.Value

	c.WriteTimeout = time.Duration(cc.WriteTimeout.Value)
	c.ReadTimeout = time.Duration(cc.ReadTimeout.Value)

  switch cc.Delimiter.Value {
	case "buffer":
		c.Delimiter = 0
	case "protocol":
		c.Delimiter = 1
	default:
		return nil, errors.New("Invalid delimiter value")
	}

	switch cc.Encoder.Value {
	case "sequence":
		c.Encoder = &SeqEncoder{}
	default:
		return nil, errors.New("Invalid encoder value")
	}

	switch cc.GetDelay.Value {
	case "none":
	default:
		return nil, errors.New("Invalid delay function")
	}

	if ch, err := MakeChannel(c); err != nil {
		return nil, err
	} else {
		return ch, nil
	}
}
