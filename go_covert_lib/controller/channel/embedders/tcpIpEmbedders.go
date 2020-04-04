package embedders

import (
	"errors"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
	"math/rand"
	"time"
	"sort"
)

// We make the fields public to facilitate logging
type TcpIpPacket struct {
	Ipv4h ipv4.Header
	Tcph  layers.TCP
	Time  time.Time
}

type TcpIpEmbedder interface {
	GetByte(p TcpIpPacket, state State) ([]byte, State, error)
	SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error)
	GetMask() [][]byte
}

// The default TcpEncoder that hides the covert message in the
// sequence number
type TcpIpSeqEncoder struct{}

// Since this function explicitely modifies the sequence number, we must ensure
// we generate a random one in the same way as createPacket
// Other implementations of this function may use other headers, and so would
// not be required to duplicate this
func (s *TcpIpSeqEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	p.Tcph.Seq = (r.Uint32() & 0xFFFFFF00) | uint32(buf[0])

	// The tcp SYN covert channel must have a change tcp sequence numbers between each packet
	// We make keep track of current and previous sequence numbers here so that
	// SetByte only needs to be called once
	if state.StoredData == nil {
	} else if seq, ok := state.StoredData.(uint32); ok {
		for seq == p.Tcph.Seq {
			p.Tcph.Seq = (r.Uint32() & 0xFFFFFF00) | uint32(buf[0])
		}
	} else {
		return p, nil, 0, state, errors.New("State has wrong data type in Stored Data: want uint32")
	}
	state.StoredData = p.Tcph.Seq
	return p, buf[1:], 0, state, nil
}

// Retrieve the byte from a packet encoded by the sequence number Encoder
func (s *TcpIpSeqEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	return []byte{byte(0xFF & p.Tcph.Seq)}, state, nil
}

func (s *TcpIpSeqEncoder) GetMask() [][]byte {
	return [][]byte{[]byte{0xFF}}
}

// Encoder stores one byte per packet in the lowest order byte of the IPV4 header ID
type TcpIpIDEncoder struct {
	emb IDEncoder
}

func (e *TcpIpIDEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newipv4h, err := e.emb.SetByte(p.Ipv4h, buf[0]); err == nil {
		p.Ipv4h = newipv4h
		return p, buf[1:], 0, state, nil
	} else {
		return p, buf, 0, state, err
	}
}

func (e *TcpIpIDEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(p.Ipv4h); err == nil {
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

func (e *TcpIpUrgPtrEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newtcph, err := e.emb.SetByte(p.Tcph, buf[0]); err == nil {
		p.Tcph = newtcph
		return p, buf[1:], 0, state, nil
	} else {
		return p, buf, 0, state, err
	}
}

func (e *TcpIpUrgPtrEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(p.Tcph); err == nil {
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

func (e *TcpIpUrgFlgEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newtcph, err := e.emb.SetByte(p.Tcph, buf[0]); err == nil {
		p.Tcph = newtcph
		return p, buf[1:], 0, state, nil
	} else {
		return p, buf, 0, state, err
	}
}

func (e *TcpIpUrgFlgEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(p.Tcph); err == nil {
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

func (e *TcpIpTimestampEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newtcph, delay, err := e.emb.SetByte(p.Tcph, buf[0]); err == nil {
		p.Tcph = newtcph
		return p, buf[1:], delay, state, nil
	} else {
		return p, buf, 0, state, err
	}
}

func (e *TcpIpTimestampEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(p.Tcph); err == nil {
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

func (e *TcpIpEcnEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}
	if newipv4h, err := e.emb.SetByte(p.Ipv4h, buf[0]); err == nil {
		p.Ipv4h = newipv4h
		return p, buf[1:], 0, state, nil
	} else {
		return p, buf, 0, state, err
	}
}

func (e *TcpIpEcnEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	if b, err := e.emb.GetByte(p.Ipv4h); err == nil {
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

func (e *TcpIpTemporalEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}

	// The first packet we send is empty, it is used to generate the initial time on the receiver
	if state.StoredData == nil {
		state.StoredData = true
		return p, buf, 0, state, nil
	} else if delay, err := e.Emb.SetByte(buf[0]); err == nil {
		return p, buf[1:], delay, state, nil
	} else {
		return p, buf, 0, state, err
	}
}

func (e *TcpIpTemporalEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	// We ignore the first packet, since it is used for setting the initial time
	if state.StoredData == nil {
		state.StoredData = p.Time
		return []byte{}, state, nil
	} else if t, ok := state.StoredData.(time.Time); ok {
		if b, err := e.Emb.GetByte(p.Time.Sub(t)); err == nil {
			state.StoredData = p.Time
			return []byte{b}, state, nil
		} else {
			return nil, state, err
		}
	} else {
		return nil, state, errors.New("State Stored Data is invalid type")
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
func (e *TcpIpEcnTempEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
	}

	// The first packet we send is empty, it is used to generate the initial time on the receiver
	if state.StoredData == nil {
		state.StoredData = true
		return p, buf, 0, state, nil
	} else {
		if delay, err := e.TmpEmb.SetByte(buf[0] & 0x01); err == nil {
			if newipv4h, err := e.ecnEmb.SetByte(p.Ipv4h, (buf[0]&0x02)>>1); err == nil {
				p.Ipv4h = newipv4h
				return p, buf[1:], delay, state, nil
			} else {
				return p, buf, 0, state, err
			}
		} else {
			return p, buf, 0, state, err
		}
	}
}

func (e *TcpIpEcnTempEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	// We ignore the first packet, since it is used for setting the initial time
	if state.StoredData == nil {
		state.StoredData = p.Time
		return []byte{}, state, nil
	} else if t, ok := state.StoredData.(time.Time); ok {
		if b1, err := e.TmpEmb.GetByte(p.Time.Sub(t)); err == nil {
			if b2, err := e.ecnEmb.GetByte(p.Ipv4h); err == nil {
				state.StoredData = p.Time
				return []byte{b1 | (b2 << 1)}, state, nil
			} else {
				return nil, state, err
			}
		} else {
			return nil, state, err
		}
	} else {
		return nil, state, errors.New("State Stored Data is invalid type")
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
func (e *TcpIpFreqEncoder) SetByte(p TcpIpPacket, buf []byte, state State) (TcpIpPacket, []byte, time.Duration, State, error) {
	if len(buf) == 0 {
		return p, nil, 0, state, errors.New("Cannot set byte if no data")
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

		return p, buf, deltaMillis, state, nil
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
		return p, buf, tdiff, state, nil
	} else {
		return p, nil, 0, state, errors.New("Incorrect state StoredData type")
	}
}

func (e *TcpIpFreqEncoder) GetByte(p TcpIpPacket, state State) ([]byte, State, error) {
	// We ignore the first packet, since it is used for setting the initial time
	if state.StoredData == nil {
		state.StoredData = freqData{ numRecv : 1, startTime : p.Time }
		return []byte{}, state, nil
	} else if fd, ok := state.StoredData.(freqData); ok {
		if p.Time.Sub(fd.startTime) >= time.Millisecond * 50 {
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
