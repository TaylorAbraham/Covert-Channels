package checksum

import (
	"bytes"
	"hash/crc32"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	encodeDecode(t, Checksum{table: crc32.MakeTable(crc32.IEEE)}, []byte{1})
	encodeDecode(t, Checksum{table: crc32.MakeTable(crc32.IEEE)}, []byte{1, 2, 3, 4, 5})
	encodeDecode(t, Checksum{table: crc32.MakeTable(crc32.Castagnoli)}, []byte{1, 2, 3, 4, 5})
	encodeDecode(t, Checksum{table: crc32.MakeTable(crc32.Koopman)}, []byte{1, 2, 3, 4, 5})
	encodeDecode(t, Checksum{table: crc32.MakeTable(crc32.IEEE)}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 100})
	encodeDecode(t, Checksum{table: crc32.MakeTable(crc32.IEEE)}, []byte{})
}

func encodeDecode(t *testing.T, c Checksum, b []byte) {
	var bcopy []byte = make([]byte, len(b))
	copy(bcopy, b)

	b2, err := c.Process(b)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if !bytes.Equal(bcopy, b) {
		t.Errorf("Original array changed")
	}
	if !bytes.Equal(b2[:len(b)], b) {
		t.Errorf("Processed leading bytes changed")
	}
	if len(b2) != len(b)+4 {
		t.Errorf("Checksum bytes not added")
	}

	bcopy = make([]byte, len(b2))
	copy(bcopy, b2)

	b3, err := c.Unprocess(b2)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if !bytes.Equal(bcopy, b2) {
		t.Errorf("Original array changed")
	}
	if !bytes.Equal(b, b3) {
		t.Errorf("Original array not restored on decode")
	}
}

func TestChangeBytes(t *testing.T) {
	runChangeTest(t, []byte{1, 2, 3, 4, 5, 6}, func(b []byte) []byte { b = append(b[:3], b[4:]...); return b })
	runChangeTest(t, []byte{1, 2, 3, 4, 5, 6}, func(b []byte) []byte { b[2] = b[2] & 0x80; return b })
}

func runChangeTest(t *testing.T, input []byte, f func([]byte) []byte) {

	c := Checksum{table: crc32.MakeTable(crc32.IEEE)}
	b, err := c.Process([]byte{1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	// Drop one byte
	b2 := f(b)

	b3, err := c.Unprocess(b2)
	if err == nil {
		t.Errorf("err = nil; want error")
	} else if err.Error() != "Checksum failure" {
		t.Errorf("err = '%s'; want 'Checksum failure'", err.Error())
	}
	if b3 != nil {
		t.Errorf("Returned non-nil byte slice from Unprocess")
	}
}

func TestUnprocessTooShort(t *testing.T) {
	c := Checksum{table: crc32.MakeTable(crc32.IEEE)}
	b, err := c.Unprocess([]byte{1, 2, 3})
	if err == nil {
		t.Errorf("err = nil; want error")
	} else if err.Error() != "Insufficient length for checksum" {
		t.Errorf("err = '%s'; want 'Checksum failure'", err.Error())
	}
	if b != nil {
		t.Errorf("Returned non-nil byte slice from Unprocess")
	}
}
