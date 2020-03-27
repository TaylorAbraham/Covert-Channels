package embedders

import (
	"time"
)

type TemporalEncoder struct {
	Midpoint time.Duration
}

func (id *TemporalEncoder) GetByte(t time.Duration) (byte, error) {
	if id.Midpoint < t {
		return 1, nil
	} else {
		return 0, nil
	}
}
func (id *TemporalEncoder) SetByte(b byte) (time.Duration, error) {
	if b != 0 {
		return time.Duration(id.Midpoint + (id.Midpoint / 2)), nil
	} else {
		return time.Duration(id.Midpoint / 2), nil
	}
}
