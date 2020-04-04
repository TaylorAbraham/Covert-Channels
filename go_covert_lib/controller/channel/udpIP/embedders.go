package udpIP

import (
	"../embedders"
	"errors"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

type UdpEncoder interface {
	GetByte(ipv4h ipv4.Header, udph layers.UDP, state embedders.State) ([]byte, embedders.State, error)
	SetByte(ipv4h ipv4.Header, udph layers.UDP, buf []byte, state embedders.State) (ipv4.Header, layers.UDP, []byte, embedders.State, error)
	GetMask() [][]byte
}

// Encoder stores one byte per packet in the lowest order byte of the IPV4 header ID
type IDEncoder struct {
	emb *embedders.IDEncoder
}

func (e *IDEncoder) GetByte(ipv4h ipv4.Header, udph layers.UDP, state embedders.State) ([]byte, embedders.State, error) {
	if b, err := e.emb.GetByte(ipv4h); err == nil {
		return []byte{b}, state, nil
	} else {
		return nil, state, err
	}
}

func (e *IDEncoder) SetByte(ipv4h ipv4.Header, udph layers.UDP, buf []byte, state embedders.State) (ipv4.Header, layers.UDP, []byte, embedders.State, error) {
	if len(buf) == 0 {
		return ipv4h, udph, nil, state, errors.New("Cannot set byte if no data")
	}
	if newipv4h, err := e.emb.SetByte(ipv4h, buf[0]); err == nil {
		return newipv4h, udph, buf[1:], state, nil
	} else {
		return ipv4h, udph, buf, state, err
	}
}

func (s *IDEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0xFF}}
}
