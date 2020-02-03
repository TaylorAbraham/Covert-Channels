package checksum

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

type Checksum struct {
	// A table with lookup values for the crc32 polynomial.
	// Created with hash/crc32.MakeTable
	table *crc32.Table
}

// Process the incoming data by adding a crc32 checksum to the end.
func (cs *Checksum) Process(data []byte) ([]byte, error) {
	check := crc32.Checksum(data, cs.table)
	newData := make([]byte, len(data)+4)
	// It's not strictly necessary to copy that data.
	// I find that the functional programming style of
	// not modifying the inputs can be safer, albeit slower.
	copy(newData, data)
	// Append the checksum
	binary.BigEndian.PutUint32(newData[len(data):], check)
	return newData, nil
}

// Validate the incoming data by interpreting the
// final four bytes as a crc32 checksum, calculating
// the actual checksum on the data, and determining if
// they match.
// Returns errors if the data length is too short
// or if the crc32 checksums don't match.
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
	// It's not strictly necessary to copy that data.
	// I find that the functional programming style of
	// not modifying the inputs can be safer, albeit slower.
	copy(newData, data[:len(data)-4])
	return newData, nil
}
