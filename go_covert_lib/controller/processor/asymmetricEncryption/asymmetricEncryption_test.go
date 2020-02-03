package asymmetricEncryption

import (
	"bytes"
	"reflect"
	"testing"
)

func TestRSAEncodeDecode(t *testing.T) {
	encodeDecode(t, make([]byte, 32), make([]byte, 32), make([]byte, 32), make([]byte, 32), []byte{1, 2, 3, 4, 5})
}

func encodeDecode(t *testing.T, senderPublicKey []byte, senderPrivateKey []byte, receiverPublicKey []byte, receiverPrivateKey []byte, b []byte) {

	cc := GetDefault()
	cc.SenderPublicKey = senderPublicKey
	cc.SenderPrivateKey = senderPrivateKey
	cc.ReceiverPublicKey = receiverPublicKey
	cc.ReceiverPrivateKey = receiverPrivateKey

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