package tcpHandshake

import (
	"../embedders"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"time"
)

type TcpEncoder interface {
	GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state embedders.State) ([]byte, embedders.State, error)
	SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state embedders.State) (ipv4.Header, layers.TCP, []byte, time.Duration, embedders.State, error)
	GetMask() [][]byte
}
