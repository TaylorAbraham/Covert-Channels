package zLibCompression

import (
	"bytes"
	"fmt"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	compressDecompress(t, ZLibCompression{}, []byte{1, 2, 3, 4, 5}, []byte{1, 2, 3, 4, 5})
	compressDecompress(t, ZLibCompression{}, []byte{'t', 'e', 's', 't'}, []byte{'t', 'e', 's', 't'})
	//compressDecompress(t, ZLibCompression{}, []byte{}, []byte{}) (*TODO create test case to handle no compression)
}

func compressDecompress(t *testing.T, z ZLibCompression, b, expected []byte) {
	var bcopy []byte = make([]byte, len(b))
	copy(bcopy, b)
	fmt.Println(bcopy)

	b2, err := z.Process(b)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if bytes.Equal(bcopy, b2) {
		t.Errorf("Original array changed")
	}

	copy(bcopy, b2)
	fmt.Println(bcopy)

	b3, err := z.Unprocess(b2)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if bytes.Equal(bcopy, b3) {
		t.Errorf("Original array changed")
	}
	if !bytes.Equal(b, b3) {
		t.Errorf("Original array not restored on decompress")
	}
}
