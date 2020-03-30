package embedders

import (
	"errors"
	"math/bits"
	"strconv"
)

type State struct {
	PacketNumber uint64
	ByteLog      []byte
	MaskSize     int
	MaskIndex    int
	StoredData   interface{}
}

func MakeState(mask [][]byte) State {
	return State{ ByteLog : []byte{}, MaskSize : len(mask) }
}

// Increments both the packet number and the MaskIndex
func (s State) IncrementState() State {
	s.MaskIndex += 1
	s.PacketNumber += 1
	if s.MaskIndex >= s.MaskSize {
		s.MaskIndex = 0
	}
	return s
}

func CalcSize(mask [][]byte, n int) (int, []byte, int, error) {
	maskArr, bitSum := toMaskArr(mask)
	if bitSum%8 != 0 {
		return 0, nil, 0, errors.New("Encoder must have bit sum as multiple of 8; found " + strconv.Itoa(bitSum))
	}
	var err error
	if 0 != (n*8)%bitSum {
		err = errors.New("Data must be multiple of " + strconv.Itoa(bitSum) + "/8 bytes")
	}
	return ((n * 8) / bitSum) * len(maskArr), maskArr, bitSum, err
}

func CalcValid(mask [][]byte, n int) (int, int, []byte, error) {
	maskArr, bitSum := toMaskArr(mask)
	if bitSum%8 != 0 {
		return 0, 0, nil, errors.New("Encoder must have bit sum as multiple of 8; found " + strconv.Itoa(bitSum))
	}
	var err error
	if 0 != n%len(maskArr) {
		err = errors.New("Data must be multiple of " + strconv.Itoa(bitSum) + " bytes")
	}
	return (n / len(maskArr)) * len(maskArr), (n / len(maskArr)) * (bitSum / 8), maskArr, err
}

func toMaskArr(mask [][]byte) ([]byte, int) {
	var (
		maskArr []byte = []byte{}
		bitSum  int    = 0
	)
	for i := range mask {
		maskArr = append(maskArr, mask[i]...)
	}
	for i := range maskArr {
		bitSum += bits.OnesCount(uint(maskArr[i]))
	}
	return maskArr, bitSum
}

func EncodeFromMask(mask [][]byte, data []byte) ([]byte, error) {

	var newData []byte

	n, maskArr, _, err := CalcSize(mask, len(data))
	if err != nil {
		return nil, err
	}

	newData = make([]byte, n)

	var (
		maskBitIndex uint = 0
		newDataIndex int  = 0
		next         bool = false
	)
	for i := range data {
		for j := 0; j < 8; j += 1 {
			next = false
			for !next {
				var maskByte byte = maskArr[newDataIndex%len(maskArr)]
				if 0 != (0x80>>maskBitIndex)&maskByte {
					if 0 != ((0x80 >> uint(j)) & data[i]) {
						newData[newDataIndex] |= 0x80 >> maskBitIndex
					}
					next = true
				}
				maskBitIndex += 1
				if maskBitIndex == 8 {
					maskBitIndex = 0
					newDataIndex += 1
				}
			}
		}
	}
	return newData, nil
}

func DecodeFromMask(mask [][]byte, data []byte) ([]byte, error) {

	_, decodeLen, maskArr, err := CalcValid(mask, len(data))
	if err != nil {
		return nil, err
	}

	var (
		newData []byte = make([]byte, decodeLen)
		newPos  uint   = 0
		newBit  uint   = 0
	)
	for len(data) > 0 {
		var subData []byte = data[:len(maskArr)]
		for i := range maskArr {
			for j := 0; j < 8; j += 1 {
				if 0 != (0x80>>uint(j))&maskArr[i] {
					if 0 != (0x80>>uint(j))&subData[i] {
						newData[newPos] |= 0x80 >> newBit
					}
					newBit += 1
					if newBit == 8 {
						newBit = 0
						newPos += 1
					}
				}
			}
		}
		data = data[len(maskArr):]
	}
	return newData, nil
}

func GetBuf(mask [][]byte, data []byte) ([]byte, error) {
	var (
		dataBufSize int
		sizeErr     error
	)
	// We must expand out the input storage array to
	// the correct size to potentially handle variable size inputs
	dataBufSize, _, _, sizeErr = CalcSize(mask, len(data))
	if sizeErr != nil {
		return nil, sizeErr
	}
	return make([]byte, dataBufSize), nil
}

func CopyData(mask [][]byte, n uint64, dataBuf, data []byte, err error) (uint64, error) {

	var sizeErr, sizeErr2 error
	var validSize int

	// Calculate how many bytes that were received correspond to
	// valid bytes (i.e. bytes that could be parsed evenly to a set of final bytes
	// using the encoder mask)
	validSize, _, _, sizeErr = CalcValid(mask, int(n))
	dataBuf = dataBuf[:validSize]
	// We decode as much as possible taking into account potentially missing bytes
	dataBuf, sizeErr2 = DecodeFromMask(mask, dataBuf)

	n = uint64(len(dataBuf))
	copy(data, dataBuf)

	if err != nil {
		return n, err
	} else if sizeErr != nil {
		return n, sizeErr
	} else if sizeErr2 != nil {
		return n, sizeErr2
	} else {
		return n, nil
	}
}

func GetSentSize(mask [][]byte, n uint64, err error) (uint64, error) {
	_, decodeLen, _, sizeErr := CalcValid(mask, int(n))
	n = uint64(decodeLen)
	if err != nil {
		return n, err
	} else if sizeErr != nil {
		return n, sizeErr
	} else {
		return n, nil
	}
}
