package checksum

import (
	"hash/crc32"
	"encoding/binary"
	"errors"
)

type Checksum struct{
	table *crc32.Table
}

func (cs *Checksum) Process(data []byte) ([]byte, error) {
	check := crc32.Checksum(data, cs.table)
	newData := make([]byte, len(data) + 4)
	copy(newData, data)
	binary.BigEndian.PutUint32(newData[len(data):], check)
	return newData, nil
}

func (cs *Checksum) Unprocess(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("Insufficient length for checksum")
	}
	check := crc32.Checksum(data[:len(data)-4], cs.table)
	inputCheck := binary.BigEndian.Uint32(data[len(data)-4:])
	if inputCheck != check {
		return nil, errors.New("Checksum failure")
	}
	newData := make([]byte, len(data)-4)
	copy(newData, data[:len(data) - 4])
	return newData, nil
}
