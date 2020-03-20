package embedders

import (
	"errors"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"math/rand"
	"time"
)

// The default TcpEncoder that hides the covert message in the
// sequence number
type TcpIpSeqEncoder struct{}

// Since this function explicitely modifies the sequence number, we must ensure
// we generate a random one in the same way as createPacket
// Other implementations of this function may use other headers, and so would
// not be required to duplicate this
func (s *TcpIpSeqEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, time.Duration, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, time.Duration(0), errors.New("Cannot set byte if no data")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tcph.Seq = (r.Uint32() & 0xFFFFFF00) | uint32(buf[0])
	return ipv4h, tcph, buf[1:], time.Duration(0), nil
}

// Retrieve the byte from a packet encoded by the sequence number Encoder
func (s *TcpIpSeqEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error) {
	return []byte{byte(0xFF & tcph.Seq)}, nil
}

// Encoder stores one byte per packet in the lowest order byte of the IPV4 header ID
type TcpIpIDEncoder struct {
	emb *IDEncoder
}

func (e *TcpIpIDEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, time.Duration, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, time.Duration(0), errors.New("Cannot set byte if no data")
	}
	if newipv4h, err := e.emb.SetByte(ipv4h, buf[0]); err == nil {
		return newipv4h, tcph, buf[1:], time.Duration(0), nil
	} else {
		return ipv4h, tcph, buf, time.Duration(0), err
	}
}

func (e *TcpIpIDEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error) {
	if b, err := e.emb.GetByte(ipv4h); err == nil {
		return []byte{b}, nil
	} else {
		return nil, err
	}
}

type TcpIpURGEncoder struct {
	emb *URGEncoder
}

func (e *TcpIpURGEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, time.Duration, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, time.Duration(0), errors.New("Cannot set byte if no data")
	}
	if newtcph, err := e.emb.SetByte(tcph, buf[0]); err == nil {
		return ipv4h, newtcph, buf[1:], time.Duration(0), nil
	} else {
		return ipv4h, tcph, buf, time.Duration(0), err
	}
}

func (e *TcpIpURGEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error) {
	if b, err := e.emb.GetByte(tcph); err == nil {
		return []byte{b}, nil
	} else {
		return nil, err
	}
}

type TcpIpTimeEncoder struct {
	emb *TimeEncoder
}

func (e *TcpIpTimeEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error) {
	if b, err := e.emb.GetByte(tcph); err == nil {
		return []byte{b}, nil
	} else {
		return nil, err
	}
}

func (e *TcpIpTimeEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, time.Duration, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, time.Duration(0), errors.New("Cannot set byte if no data")
	}
	if newtcph, delay, err := e.emb.SetByte(tcph, buf[0]); err == nil {
		return ipv4h, newtcph, buf[1:], delay, nil
	} else {
		return ipv4h, tcph, buf, time.Duration(0), err
	}
}
