package tcpSyn

import (
	"errors"
	"time"

	"../../config"
	"../embedders"
)

type ConfigClient struct {
	FriendIP     config.IPV4Param
	OriginIP     config.IPV4Param
	BounceIP     config.IPV4Param
	FriendPort   config.U16Param
	OriginPort   config.U16Param
	BouncePort   config.U16Param
	Bounce       config.BoolParam
	Delimiter    config.SelectParam
	Embedder     config.SelectParam
	WriteTimeout config.U64Param
	ReadTimeout  config.U64Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP:     config.MakeIPV4("127.0.0.1", config.Display{Description: "Your friend's IP address.", Name: "Friend's IP", Group: "IP Addresses"}),
		OriginIP:     config.MakeIPV4("127.0.0.1", config.Display{Description: "Your IP address.", Name: "Your IP", Group: "IP Addresses"}),
		FriendPort:   config.MakeU16(8123, [2]uint16{0, 65535}, config.Display{Description: "Your friend's port.", Name: "Friend's Port", Group: "Ports"}),
		OriginPort:   config.MakeU16(8124, [2]uint16{0, 65535}, config.Display{Description: "Your port.", Name: "Your Port", Group: "Ports"}),
		Bounce:       config.MakeBool(false, config.Display{Description: "Toggle bounce mode, which spoofs your IP address.", Name: "Bounce", Group: "Bouncing", GroupToggle: true}),
		BounceIP:     config.MakeIPV4("127.0.0.1", config.Display{Description: "The bouncer's IP address.", Name: "Bouncer's IP", Group: "Bouncing"}),
		BouncePort:   config.MakeU16(0, [2]uint16{0, 65535}, config.Display{Description: "The bouncer's port.", Name: "Bouncer's Port", Group: "Bouncing"}),
		WriteTimeout: config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The write timeout in milliseconds.", Name: "Write Timeout", Group: "Timing"}),
		ReadTimeout:  config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The read timeout in milliseconds.", Name: "Read Timeout", Group: "Timing"}),
		Delimiter:    config.MakeSelect("protocol", []string{"buffer", "protocol"}, config.Display{Description: "The delimiter to use for deciding when to return after having received a message.", Name: "Delimeter", Group: "Settings"}),
		Embedder:     config.MakeSelect("sequence", []string{"sequence", "id", "urgptr", "urgflg", "timestamp", "ecn"}, config.Display{Description: "The encoding mechanism to use for this protocol.", Name: "Encoding", Group: "Settings"}),
	}
}

func ToChannel(cc ConfigClient) (*Channel, error) {
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

	c.WriteTimeout = time.Duration(cc.WriteTimeout.Value) * time.Millisecond
	c.ReadTimeout = time.Duration(cc.ReadTimeout.Value) * time.Millisecond

	switch cc.Delimiter.Value {
	case "buffer":
		c.Delimiter = 0
	case "protocol":
		c.Delimiter = 1
	default:
		return nil, errors.New("Invalid delimiter value")
	}

	if cc.Bounce.Value && cc.Embedder.Value != "sequence" {
		return nil, errors.New("Only the sequence embedder is supported in bounce mode")
	}

	switch cc.Embedder.Value {
	case "sequence":
		c.Embedder = &embedders.TcpIpSeqEncoder{}
	case "id":
		c.Embedder = &embedders.TcpIpIDEncoder{}
	case "urgflg":
		c.Embedder = &embedders.TcpIpUrgFlgEncoder{}
	case "urgptr":
		c.Embedder = &embedders.TcpIpUrgPtrEncoder{}
	case "timestamp":
		c.Embedder = &embedders.TcpIpTimestampEncoder{}
	case "ecn":
		c.Embedder = &embedders.TcpIpEcnEncoder{}
	default:
		return nil, errors.New("Invalid embedder value")
	}

	if ch, err := MakeChannel(c); err != nil {
		return nil, err
	} else {
		return ch, nil
	}
}
