package symmetricEncryption

import (
	"bytes"
	"reflect"
	"testing"
)

func TestAESEncodeDecode(t *testing.T) {
	encodeDecode(t, make([]byte, 32), "Advanced Encryption Standard (AES)", "Cipher Block Chaining (CBC)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 32), "Advanced Encryption Standard (AES)", "Cipher Feedback (CFB)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 32), "Advanced Encryption Standard (AES)", "Counter (CTR)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 32), "Advanced Encryption Standard (AES)", "Output Feedback (OFB)", []byte{1, 2, 3, 4, 5})
}

func TestDESEncodeDecode(t *testing.T) {
	encodeDecode(t, make([]byte, 8), "Data Encryption Standard (DES)", "Cipher Block Chaining (CBC)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 8), "Data Encryption Standard (DES)", "Cipher Feedback (CFB)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 8), "Data Encryption Standard (DES)", "Counter (CTR)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 8), "Data Encryption Standard (DES)", "Output Feedback (OFB)", []byte{1, 2, 3, 4, 5})
}

func Test3DESEncodeDecode(t *testing.T) {
	encodeDecode(t, make([]byte, 24), "Triple Data Encryption Standard (3DES)", "Cipher Block Chaining (CBC)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 24), "Triple Data Encryption Standard (3DES)", "Cipher Feedback (CFB)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 24), "Triple Data Encryption Standard (3DES)", "Counter (CTR)", []byte{1, 2, 3, 4, 5})
	encodeDecode(t, make([]byte, 24), "Triple Data Encryption Standard (3DES)", "Output Feedback (OFB)", []byte{1, 2, 3, 4, 5})
}

func TestPadUnpad(t *testing.T) {
	b := Pad([]byte{76, 42, 98, 34, 88, 12, 45, 32, 76})
	if !reflect.DeepEqual(b, []byte{0, 0, 0, 0, 0, 0, 0, 9, 76, 42, 98, 34, 88, 12, 45, 32, 76, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Errorf("Array not padded correctly")
	}

	b = UnPad(b)
	if !reflect.DeepEqual(b, []byte{76, 42, 98, 34, 88, 12, 45, 32, 76}) {
		t.Errorf("Array not unpadded correctly")
	}
}

func encodeDecode(t *testing.T, key []byte, algo string, mode string, b []byte) {

	cc := GetDefault()
	cc.Key.Value = key
	cc.Algorithm.Value = algo
	cc.Mode.Value = mode

	c, err := ToProcessor(cc)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	var bcopy []byte = make([]byte, len(b))
	copy(bcopy, b)

	b2, err := c.Process(b)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}

	if !bytes.Equal(bcopy, b) {
		t.Errorf("Original array changed")
	}

	bcopy = make([]byte, len(b2))
	copy(bcopy, b2)

	b3, err := c.Unprocess(b2)
	if err != nil {
		t.Errorf("err = '%s'; want nil", err.Error())
	}
	//if !bytes.Equal(bcopy, b2) {
	//	t.Errorf("Original array changed")
	//}
	if !bytes.Equal(b, b3) {
		t.Errorf("Original array not restored on decode")
	}
}
