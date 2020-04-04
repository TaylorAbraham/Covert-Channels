package tcpSyn

import (
	"../embedders"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"time"
)

// An encoded may be provided to the TCP Channel
// to configure where the covert message is hidden
// A method is used to hide the byte during sending, and a second is used
// extract it during reception
type TcpEncoder interface {
	GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state embedders.State) ([]byte, embedders.State, error)
	SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state embedders.State) (ipv4.Header, layers.TCP, []byte, time.Duration, embedders.State, error)
	GetMask() [][]byte
}
