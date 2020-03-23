package embedders

import (
	"testing"
	"bytes"
)

func TestEncode(t *testing.T) {
	runTest(t,[][]byte{[]byte{0xFF}}, []byte{1,2,3,4,5}, []byte{1,2,3,4,5})
	runTest(t,[][]byte{[]byte{0x0F, 0x0F}}, []byte{1,2,3,4,5}, []byte{0, 1, 0, 2, 0, 3, 0, 4, 0, 5})
	runTest(t,[][]byte{[]byte{0x0F}, []byte{0x0F}}, []byte{1,2,3,4,5}, []byte{0, 1, 0, 2, 0, 3, 0, 4, 0, 5})
	runTest(t,[][]byte{[]byte{0x0F}, []byte{0xF0}}, []byte{1,2,3,4,5}, []byte{0, 0x10, 0, 0x20, 0, 0x30, 0, 0x40, 0, 0x50})
	runTest(t,[][]byte{[]byte{0x01}, []byte{0x7F}}, []byte{1,2,3,4,5}, []byte{0, 1, 0, 2, 0, 3, 0, 4, 0, 5})
	runTest(t,[][]byte{[]byte{0x3F}, []byte{0x03}}, []byte{1,2,3,4,5}, []byte{0, 1, 0, 2, 0, 3, 1, 0, 1, 1})

	runTest(t,[][]byte{[]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}}, []byte{1,2,3,4,5},
		[]byte{0, 0, 0, 0, 0, 0, 0, 1,
					 0, 0, 0, 0, 0, 0, 1, 0,
					 0, 0, 0, 0, 0, 0, 1, 1,
					 0, 0, 0, 0, 0, 1, 0, 0,
					 0, 0, 0, 0, 0, 1, 0, 1})

	runTest(t,
					[][]byte{[]byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01},
									 []byte{0x01}, []byte{0x01}, []byte{0x01}, []byte{0x01}},
					[]byte{1,2,3,4,5},
					[]byte{0, 0, 0, 0, 0, 0, 0, 1,
						 		 0, 0, 0, 0, 0, 0, 1, 0,
						 		 0, 0, 0, 0, 0, 0, 1, 1,
						 		 0, 0, 0, 0, 0, 1, 0, 0,
						 		 0, 0, 0, 0, 0, 1, 0, 1})
}

func runTest(t *testing.T, mask [][]byte, input []byte, expected []byte) {
	output, err := EncodeFromMask(mask, input)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	if bytes.Compare(expected, output) != 0 {
		t.Errorf("Encoding expected = %v; got %v", expected, output)
	}
	output2, err2 := DecodeFromMask(mask, output)
	if err2 != nil {
		t.Errorf("err = '%s'; want nil", err2.Error())
	}
	if bytes.Compare(input, output2) != 0 {
		t.Errorf("Decoding expected = %v; got %v", input, output2)
	}
}
