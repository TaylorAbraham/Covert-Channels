package udpIP

import (
	"../embedders"
	"errors"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

type UdpEncoder interface {
	GetByte(ipv4h ipv4.Header, udph layers.UDP, maskIndex int) ([]byte, error)
	SetByte(ipv4h ipv4.Header, udph layers.UDP, buf []byte, maskIndex int) (ipv4.Header, layers.UDP, []byte, error)
	GetMask() [][]byte
}

// Encoder stores one byte per packet in the lowest order byte of the IPV4 header ID
type IDEncoder struct {
	emb *embedders.IDEncoder
}

func (e *IDEncoder) GetByte(ipv4h ipv4.Header, udph layers.UDP, maskIndex int) ([]byte, error) {
	if b, err := e.emb.GetByte(ipv4h); err == nil {
		return []byte{b}, nil
	} else {
		return nil, err
	}
}
func (e *IDEncoder) SetByte(ipv4h ipv4.Header, udph layers.UDP, buf []byte, maskIndex int) (ipv4.Header, layers.UDP, []byte, error) {
	if len(buf) == 0 {
		return ipv4h, udph, nil, errors.New("Cannot set byte if no data")
	}
	if newipv4h, err := e.emb.SetByte(ipv4h, buf[0]); err == nil {
		return newipv4h, udph, buf[1:], nil
	} else {
		return ipv4h, udph, buf, err
	}
}

func (s *IDEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0xFF}}
}
