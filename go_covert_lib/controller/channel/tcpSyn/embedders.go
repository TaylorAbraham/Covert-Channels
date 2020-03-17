package tcpSyn

import (
	"golang.org/x/net/ipv4"
	"time"
	"github.com/google/gopacket/layers"
)

// An encoded may be provided to the TCP Channel
// to configure where the covert message is hidden
// A method is used to hide the byte during sending, and a second is used
// extract it during reception
type TcpEncoder interface {
	// We use the prevSequence to indicate duplicate packets arriving
	// on from the bouncer socket. It is this function's responsibility
	// to ensure that any modifications to the tcp header ensure that the
	// new sequence number is different than the previous sequence number
	SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte) (ipv4.Header, layers.TCP, []byte, time.Duration, error)
	GetByte(ipv4h ipv4.Header, tcph layers.TCP) ([]byte, error)
}
