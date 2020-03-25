package embedders

import (
	"golang.org/x/net/ipv4"
	"math/rand"
	"time"
)

type IDEncoder struct{}

func (id *IDEncoder) GetByte(ipv4h ipv4.Header) (byte, error) {
	return byte(ipv4h.ID & 0xFF), nil
}

func (id *IDEncoder) SetByte(ipv4h ipv4.Header, b byte) (ipv4.Header, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ipv4h.ID = (r.Int() & 0xFF00) | int(b)
	// Based on my experimental results, the raw socket will override
	// an IP ID of zero. We use this loop to ensure that the ID is something
	// other than zero so that our real data is transmitted
	for ipv4h.ID == 0 {
		ipv4h.ID = (r.Int() & 0xFF00) | int(b)
	}

	return ipv4h, nil
}

type EcnEncoder struct{}

func (id *EcnEncoder) GetByte(ipv4h ipv4.Header) (byte, error) {
	if ipv4h.TOS&0x02 != 0 {
		return 1, nil
	} else {
		return 0, nil
	}
}

func (id *EcnEncoder) SetByte(ipv4h ipv4.Header, b byte) (ipv4.Header, error) {

	if b != 0 {
		ipv4h.TOS = (ipv4h.TOS & 0xFC) | 0x03
	} else {
		ipv4h.TOS = (ipv4h.TOS & 0xFC) | 0x01
	}

	return ipv4h, nil
}
