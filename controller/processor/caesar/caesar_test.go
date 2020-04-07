package caesar

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	encodeDecode(t, Caesar{Shift: 0}, []byte{1, 2, 3, 4, 5}, []byte{1, 2, 3, 4, 5})
	encodeDecode(t, Caesar{Shift: 5}, []byte{1, 2, 3, 4, 5}, []byte{6, 7, 8, 9, 10})
	encodeDecode(t, Caesar{Shift: -10}, []byte{1, 2, 3, 4, 5}, []byte{247, 248, 249, 250, 251})
	encodeDecode(t, Caesar{Shift: 25}, []byte{}, []byte{})
}

func encodeDecode(t *testing.T, c Caesar, b, expected []byte) {
	var bcopy []byte = make([]byte, len(b))
	copy(bcopy, b)

	b2, err := c.Process(b)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if !bytes.Equal(bcopy, b) {
		t.Errorf("Original array changed")
	}

	for i := range b2 {
		if b2[i] != expected[i] {
			t.Errorf("Byte %d not shifted: got %d expected %d", i, b2[i], expected[i])
		}
	}

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
