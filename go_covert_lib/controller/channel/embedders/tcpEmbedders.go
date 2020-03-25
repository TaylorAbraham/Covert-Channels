package embedders

import (
	"encoding/binary"
	"errors"
	"github.com/google/gopacket/layers"
	"time"
)

type UrgPtrEncoder struct{}

func (id *UrgPtrEncoder) GetByte(tcph layers.TCP) (byte, error) {
	return byte(tcph.Urgent & 0xFF), nil
}
func (id *UrgPtrEncoder) SetByte(tcph layers.TCP, b byte) (layers.TCP, error) {
	tcph.Urgent = uint16(b)
	return tcph, nil
}

type UrgFlgEncoder struct{}

func (id *UrgFlgEncoder) GetByte(tcph layers.TCP) (byte, error) {
	if tcph.URG {
		return 1, nil
	} else {
		return 0, nil
	}
}
func (id *UrgFlgEncoder) SetByte(tcph layers.TCP, b byte) (layers.TCP, error) {
	tcph.URG = b != 0
	return tcph, nil
}

type TimeEncoder struct{}

func (id *TimeEncoder) GetByte(tcph layers.TCP) (byte, error) {
	for i := range tcph.Options {
		if layers.TCPOptionKindTimestamps == tcph.Options[i].OptionType {
			if 4 <= len(tcph.Options[i].OptionData) {
				return tcph.Options[i].OptionData[3], nil
			} else {
				return 0, errors.New("Option data too short")
			}
		}
	}
	return 0, errors.New("Missing timestamp option")
}

func (id *TimeEncoder) SetByte(tcph layers.TCP, b byte) (layers.TCP, time.Duration, error) {
	var opt layers.TCPOption
	opt.OptionType = layers.TCPOptionKindTimestamps
	opt.OptionLength = 10
	opt.OptionData = make([]byte, 8)

	var optIndex int = 0
	var delay time.Duration = 0

	for _ = range tcph.Options {
		if tcph.Options[optIndex].OptionType == layers.TCPOptionKindTimestamps {
			break
		}
		optIndex += 1
	}
	if optIndex != len(tcph.Options) {

		copy(opt.OptionData, tcph.Options[optIndex].OptionData)

		var currTime uint32 = binary.BigEndian.Uint32(opt.OptionData[:4])
		opt.OptionData[3] = b
		var newTime uint32 = binary.BigEndian.Uint32(opt.OptionData[:4])
		if newTime < currTime {
			newTime += 256
		}

		binary.BigEndian.PutUint32(opt.OptionData[:4], newTime)

		delay = time.Duration(newTime-currTime) * time.Millisecond
		tcph.Options[optIndex] = opt
	} else {
		binary.BigEndian.PutUint32(opt.OptionData[:4], uint32(time.Now().UnixNano()/10000000))
		opt.OptionData[3] = b
		tcph.Options = append(tcph.Options, opt)
	}

	return tcph, delay, nil
}
