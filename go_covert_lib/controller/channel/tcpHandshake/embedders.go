package tcpHandshake

import (
	"golang.org/x/net/ipv4"
	"time"
	"github.com/google/gopacket/layers"
)

type TcpEncoder interface {
	GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error)
	SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, time.Duration, error)
}
