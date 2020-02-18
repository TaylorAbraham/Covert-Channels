package gZipCompression

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	compressDecompress(t, GZipCompression{}, []byte{1, 2, 3, 4, 5}, []byte{1, 2, 3, 4, 5})
	compressDecompress(t, GZipCompression{}, []byte{'t', 'e', 's', 't'}, []byte{'t', 'e', 's', 't'})
	//compressDecompress(t, GZipCompression{}, []byte{}, []byte{}) (*TODO create test case to handle no compression)
}

func compressDecompress(t *testing.T, g GZipCompression, b, expected []byte) {
	var bcopy []byte = make([]byte, len(b))
	copy(bcopy, b)

	b2, err := g.Process(b)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if bytes.Equal(bcopy, b2) {
		t.Errorf("Original array changed")
	}

	copy(bcopy, b2)

	b3, err := g.Unprocess(b2)
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
