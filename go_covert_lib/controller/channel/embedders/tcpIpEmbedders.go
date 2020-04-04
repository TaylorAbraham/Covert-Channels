package embedders

import (
	"errors"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"math/rand"
	"time"
	"sort"
)

// The default TcpEncoder that hides the covert message in the
// sequence number
type TcpIpSeqEncoder struct{}

// Since this function explicitely modifies the sequence number, we must ensure
// we generate a random one in the same way as createPacket
// Other implementations of this function may use other headers, and so would
// not be required to duplicate this
func (s *TcpIpSeqEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tcph.Seq = (r.Uint32() & 0xFFFFFF00) | uint32(buf[0])

	// The tcp SYN covert channel must have a change tcp sequence numbers between each packet
	// We make keep track of current and previous sequence numbers here so that
	// SetByte only needs to be called once
	if state.StoredData == nil {
	} else if seq, ok := state.StoredData.(uint32); ok {
		for seq == tcph.Seq {
			tcph.Seq = (r.Uint32() & 0xFFFFFF00) | uint32(buf[0])
		}
	} else {
		return ipv4h, tcph, nil, 0, state, errors.New("State has wrong data type in Stored Data: want uint32")
	}
	state.StoredData = tcph.Seq
	return ipv4h, tcph, buf[1:], 0, state, nil
}

// Retrieve the byte from a packet encoded by the sequence number Encoder
func (s *TcpIpSeqEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	return []byte{byte(0xFF & tcph.Seq)}, state, nil
}

func (s *TcpIpSeqEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0xFF}}
}

// Encoder stores one byte per packet in the lowest order byte of the IPV4 header ID
type TcpIpIDEncoder struct {
	emb IDEncoder
}

func (e *TcpIpIDEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newipv4h, err := e.emb.SetByte(ipv4h, buf[0]); err == nil {
		return newipv4h, tcph, buf[1:], 0, state, nil
	} else {
		return ipv4h, tcph, buf, 0, state, err
	}
}

func (e *TcpIpIDEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(ipv4h); err == nil {
		return []byte{b}, state, nil
	} else {
		return nil, state, err
	}
}

func (s *TcpIpIDEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0xFF}}
}

type TcpIpUrgPtrEncoder struct {
	emb UrgPtrEncoder
}

func (e *TcpIpUrgPtrEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newtcph, err := e.emb.SetByte(tcph, buf[0]); err == nil {
		return ipv4h, newtcph, buf[1:], 0, state, nil
	} else {
		return ipv4h, tcph, buf, 0, state, err
	}
}

func (e *TcpIpUrgPtrEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(tcph); err == nil {
		return []byte{b}, state, nil
	} else {
		return nil, state, err
	}
}

func (s *TcpIpUrgPtrEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0xFF}}
}

type TcpIpUrgFlgEncoder struct {
	emb UrgFlgEncoder
}

func (e *TcpIpUrgFlgEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newtcph, err := e.emb.SetByte(tcph, buf[0]); err == nil {
		return ipv4h, newtcph, buf[1:], 0, state, nil
	} else {
		return ipv4h, tcph, buf, 0, state, err
	}
}

func (e *TcpIpUrgFlgEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(tcph); err == nil {
		return []byte{b}, state, nil
	} else {
		return nil, state, err
	}
}

func (s *TcpIpUrgFlgEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01},
		[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01}}
}

type TcpIpTimestampEncoder struct {
	emb TimestampEncoder
}

func (e *TcpIpTimestampEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newtcph, delay, err := e.emb.SetByte(tcph, buf[0]); err == nil {
		return ipv4h, newtcph, buf[1:], delay, state, nil
	} else {
		return ipv4h, tcph, buf, 0, state, err
	}
}

func (e *TcpIpTimestampEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(tcph); err == nil {
		return []byte{b}, state, nil
	} else {
		return nil, state, err
	}
}

func (s *TcpIpTimestampEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0xFF}}
}

type TcpIpEcnEncoder struct {
	emb EcnEncoder
}

func (e *TcpIpEcnEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if ipv4h, err := e.emb.SetByte(ipv4h, buf[0]); err == nil {
		return ipv4h, tcph, buf[1:], 0, state, nil
	} else {
		return ipv4h, tcph, buf, 0, state, err
	}
}

func (e *TcpIpEcnEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(ipv4h); err == nil {
		return []byte{b}, state, nil
	} else {
		return nil, state, err
	}
}

func (s *TcpIpEcnEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01},
		[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01}}
}

type TcpIpTemporalEncoder struct {
	Emb TemporalEncoder
}

func (e *TcpIpTemporalEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}

	// The first packet we send is empty, it is used to generate the initial time on the receiver
	if state.StoredData == nil {
		state.StoredData = true
		return ipv4h, tcph, buf, 0, state, nil
	} else if t, err := e.Emb.SetByte(buf[0]); err == nil {
		return ipv4h, tcph, buf[1:], t, state, nil
	} else {
		return ipv4h, tcph, buf, 0, state, err
	}
}

func (e *TcpIpTemporalEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	// We ignore the first packet, since it is used for setting the initial time
	if state.StoredData == nil {
		state.StoredData = true
		return []byte{}, state, nil
	} else if b, err := e.Emb.GetByte(t); err == nil {
		return []byte{b}, state, nil
	} else {
		return nil, state, err
	}
}

func (s *TcpIpTemporalEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01},
		[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01}}
}

type TcpIpEcnTempEncoder struct {
	TmpEmb TemporalEncoder
	ecnEmb *EcnEncoder
}

// Encode even number bits using time delays
// Encode odd number bits using ecn flags
func (e *TcpIpEcnTempEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}

	// The first packet we send is empty, it is used to generate the initial time on the receiver
	if state.StoredData == nil {
		state.StoredData = true
		return ipv4h, tcph, buf, 0, state, nil
	} else {
		if t, err := e.TmpEmb.SetByte(buf[0] & 0x01); err == nil {
			if newipv4h, err := e.ecnEmb.SetByte(ipv4h, (buf[0]&0x02)>>1); err == nil {
				return newipv4h, tcph, buf[1:], t, state, nil
			} else {
				return ipv4h, tcph, buf, 0, state, err
			}
		} else {
			return ipv4h, tcph, buf, 0, state, err
		}
	}
}

func (e *TcpIpEcnTempEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	// We ignore the first packet, since it is used for setting the initial time
	if state.StoredData == nil {
		state.StoredData = true
		return []byte{}, state, nil
	} else {
		if b1, err := e.TmpEmb.GetByte(t); err == nil {
			if b2, err := e.ecnEmb.GetByte(ipv4h); err == nil {
				return []byte{b1 | (b2 << 1)}, state, nil
			} else {
				return nil, state, err
			}
		} else {
			return nil, state, err
		}
	}
}

func (s *TcpIpEcnTempEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0x03}, []byte{0x03}, []byte{0x03}, []byte{0x03}}
}

type TcpIpFreqEncoder struct {}

type freqData struct {
	numToSend uint
	numSent   uint
	numRecv   uint
	startTime time.Time
	sendMS    []int
}

// Encode even number bits using time delays
// Encode odd number bits using ecn flags
func (e *TcpIpFreqEncoder) SetByte(ipv4h ipv4.Header, tcph layers.TCP, buf []byte, state State) (ipv4.Header, layers.TCP, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return ipv4h, tcph, nil, 0, state, errors.New("Cannot set byte if no data")
	}

	var (
		deltaMillis time.Duration
		fd freqData
		ok bool
	)

	if state.StoredData == nil {
		var r  *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
		if buf[0] == 1 {
		 	fd = freqData{ numToSend : 8, sendMS : make([]int, 8) }
		} else {
			fd = freqData{ numToSend : 2, sendMS : make([]int, 2) }
		}
		// We send the packets at random times within the 50 ms interval
		for i := 0; i < int(fd.numToSend); i++ {
			fd.sendMS[i] = int(1 + r.Int63n(40))
		}
		sort.Ints(fd.sendMS)

		deltaMillis = time.Millisecond * time.Duration(fd.sendMS[0])
		fd.startTime = time.Now().Add(deltaMillis)
		state.StoredData = fd

		return ipv4h, tcph, buf, deltaMillis, state, nil
	} else if fd, ok = state.StoredData.(freqData); ok {
		fd.numSent += 1
		if fd.numSent < fd.numToSend {
			deltaMillis = time.Millisecond * time.Duration(fd.sendMS[fd.numSent])
			state.StoredData = fd
		} else {
			deltaMillis = time.Millisecond * 55
			state.StoredData = nil
			buf = buf[1:]
		}

		tdiff := fd.startTime.Add(deltaMillis).Sub(time.Now())
		if tdiff < 0 {
			tdiff = 0
		}
		return ipv4h, tcph, buf, tdiff, state, nil
	} else {
		return ipv4h, tcph, nil, 0, state, errors.New("Incorrect state StoredData type")
	}
}

func (e *TcpIpFreqEncoder) GetByte(ipv4h ipv4.Header, tcph layers.TCP, t time.Duration, state State) ([]byte, State, error) {
	// We ignore the first packet, since it is used for setting the initial time
	if state.StoredData == nil {
		state.StoredData = freqData{ numRecv : 1, startTime : time.Now() }
		return []byte{}, state, nil
	} else if fd, ok := state.StoredData.(freqData); ok {
		if time.Now().Sub(fd.startTime) >= time.Millisecond * 50 {
			state.StoredData = nil
			if fd.numRecv < 4 {
				return []byte{0}, state, nil
			} else {
				return []byte{1}, state, nil
			}
		} else {
			fd.numRecv += 1
			state.StoredData = fd
			return []byte{}, state, nil
		}
	} else {
		return nil, state, errors.New("Incorrect state StoredData type")
	}
}

func (s *TcpIpFreqEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01},
		[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01}}
}
