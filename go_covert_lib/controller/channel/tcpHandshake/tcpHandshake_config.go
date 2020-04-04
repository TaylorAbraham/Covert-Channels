package tcpHandshake

import (
	"../../config"
	"../embedders"
	"errors"
	"time"
)

type ConfigClient struct {
	FriendIP          config.IPV4Param
	OriginIP          config.IPV4Param
	FriendReceivePort config.U16Param
	OriginReceivePort config.U16Param
	Embedder          config.SelectParam
	DialTimeout       config.U64Param
	AcceptTimeout     config.U64Param
	ReadTimeout       config.U64Param
	WriteTimeout      config.U64Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		FriendIP:          config.MakeIPV4("127.0.0.1", config.Display{Description: "Your friend's IP address.", Name: "Friend's IP", Group: "IP Addresses"}),
		OriginIP:          config.MakeIPV4("127.0.0.1", config.Display{Description: "Your IP address.", Name: "Your IP", Group: "IP Addresses"}),
		FriendReceivePort: config.MakeU16(8123, [2]uint16{0, 65535}, config.Display{Description: "Your friend's TCP receive port. Their send port is chosen randomly.", Name: "Friend's Receive Port", Group: "Ports"}),
		OriginReceivePort: config.MakeU16(8124, [2]uint16{0, 65535}, config.Display{Description: "Your TCP receive port. Your send port is chosen randomly.", Name: "Your Receive Port", Group: "Ports"}),
		DialTimeout:       config.MakeU64(500, [2]uint64{0, 65535}, config.Display{Description: "The dial timeout for the send method in milliseconds. Zero for no timeout.", Name: "Dial Timeout", Group: "Timing"}),
		AcceptTimeout:     config.MakeU64(0, [2]uint64{0, 65535}, config.Display{Description: "The accept timeout for the receive method in milliseconds. Zero for no timeout.", Name: "Accept Timeout", Group: "Timing"}),
		ReadTimeout:       config.MakeU64(500, [2]uint64{0, 65535}, config.Display{Description: "The intra-packet read timeout for the receive method in milliseconds. Zero for no timeout.", Name: "Read Timeout", Group: "Timing"}),
		WriteTimeout:      config.MakeU64(500, [2]uint64{0, 65535}, config.Display{Description: "The a timeout for writing packets to the raw socket, in milliseconds. Zero for no timeout.", Name: "Write Timeout", Group: "Timing"}),
		Embedder:          config.MakeSelect("id", []string{"id", "urgflg", "urgptr", "timestamp", "ecn", "temporal", "frequency", "ecntemporal"}, config.Display{Description: "The encoding mechanism to use for this protocol.", Name: "Encoding", Group: "Settings"}),
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

	c.DialTimeout = time.Duration(cc.DialTimeout.Value) * time.Millisecond
	c.AcceptTimeout = time.Duration(cc.AcceptTimeout.Value) * time.Millisecond
	c.ReadTimeout = time.Duration(cc.ReadTimeout.Value) * time.Millisecond
	c.WriteTimeout = time.Duration(cc.WriteTimeout.Value) * time.Millisecond

	switch cc.Embedder.Value {
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
	case "temporal":
		c.Embedder = &embedders.TcpIpTemporalEncoder{Emb: embedders.TemporalEncoder{time.Duration(50 * time.Millisecond)}}
	case "frequency":
		c.Embedder = &embedders.TcpIpFreqEncoder{}
	case "ecntemporal":
		c.Embedder = &embedders.TcpIpEcnTempEncoder{TmpEmb: embedders.TemporalEncoder{time.Duration(50 * time.Millisecond)}}
	default:
		return nil, errors.New("Invalid embedder value")
	}

	if ch, err := MakeChannel(c); err != nil {
		return nil, err
	} else {
		return ch, nil
	}
}
