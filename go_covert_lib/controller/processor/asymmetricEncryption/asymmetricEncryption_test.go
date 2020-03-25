package asymmetricEncryption

import (
	"bytes"
	//	"reflect"
	"testing"
)

func TestRSAEncodeDecode(t *testing.T) {
	encodeDecode(t, examplePublic, examplePrivate, examplePublic, examplePrivate, []byte{1, 2, 3, 4, 5})
}

// Examples generated from https://travistidwell.com/jsencrypt/demo/
var examplePublic string = "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1+tt0BR+tOPPpc/oEgCrYtKvZPwiDgH+h9WSI5zOXZs6wnb4Hj94v//o2wAtYbiKArZDlsJt0G3ylDwVNeBeCUSPjkc0tI80pEB2czfgonE7Y5NogoWE+xQfdRwn2QOw5LD5QPkSKZ+2W8kFzwsnZrEemCXorV4Di2wcpzh7m2zGvc8NUoXbbtvR7RUhTz98NgKNnYCQlS+W2pF/aOiORjq+qqkzbedwLK/WXHGiHrt7Z/2alF+SKwLBMW5Me1TuXghl8wgxShQLu0yN3bwJMNkNMDDE1MfiPeFdDRn12B5ecrn8yZoLjXpBqpiNPX93UUgLToz+BGgypTGT1FcDWwIDAQAB\n-----END PUBLIC KEY-----"
var examplePrivate string = "-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEA1+tt0BR+tOPPpc/oEgCrYtKvZPwiDgH+h9WSI5zOXZs6wnb4Hj94v//o2wAtYbiKArZDlsJt0G3ylDwVNeBeCUSPjkc0tI80pEB2czfgonE7Y5NogoWE+xQfdRwn2QOw5LD5QPkSKZ+2W8kFzwsnZrEemCXorV4Di2wcpzh7m2zGvc8NUoXbbtvR7RUhTz98NgKNnYCQlS+W2pF/aOiORjq+qqkzbedwLK/WXHGiHrt7Z/2alF+SKwLBMW5Me1TuXghl8wgxShQLu0yN3bwJMNkNMDDE1MfiPeFdDRn12B5ecrn8yZoLjXpBqpiNPX93UUgLToz+BGgypTGT1FcDWwIDAQABAoIBAHZfWoeeBMz0q80yiv77oPn/mSqa06ysSTd8za56c+R7ip48DNDAaVmRWb5efYK6YecUtz86fmurKzc7LUGpLMSV8sHEpc9rRyfZM1b9RkioHS/9C2mq+3mO0aQpeGsQC/WEVFHbeqqZJadyMJ4Odl5lMemltsb86KKR9a9zVsiguvE6VRHJSEx6/6iXy2JHwS1FSPOT2SzrgI78TSeeROz1CqggJtqZEaAD34SVkzUKst3LvG8Z7UB+mgBzT1V39RwKx70xpBqf9GqhewH5pEOrxUozJ6hrz4sToeDpJ2ubyo7L6tsSFC7pWNciiODCJqtARBUwblOwJx0V1c72lOkCgYEA8yj2k1Zx1QxbVZb0xR56TXRVd3oMNHb9mS0OzFZfL1Bkszlihi8BK2p4gNzQJsRHzmhLND0Uvh7U1v1oF4demiwURgt2dq5bAb1/PtFhKIipdPyiKEPdcxMpEKP6i1UeotdnxomtT31YUS9omVYuiLSUOKuWU5WLZkc9wlRqp/0CgYEA41I7Ah4BeGy5cXk1Haqqo13mL5rUuOkdd5CBR9UHd9xzCxXE2z2L12/J6LcmPxsNheDOFFMLCpSglW16tkFBAOxgjpKidOpI0UG2cIiFRUTpJ8uGygcTfVxqq45jGHs89x54e9AROISweREw0mnI6N+dhMbT9JVy7x1ey9umXDcCgYAvcXeiycQOEIolif2aFFdCk4c1d4+4ENtsLplrjxKlVadAPNsXWUZ+JRj785l9ZuCnyjuaJqzMZ5GZnPnZVWVE6YLPI99qSpyhG0sfg5TUZs3BcKVm+87SbBOgFo6E7we6OBMcbrJtBwTbWkerW2Ba9fjRkdET3+LCAvZu2y+wNQKBgHF7YqPq8Nb6iBVC6iZWRftqa/iF9f4dui0vQarniWPn9LKq+mxsrDwvvX9ktz43tieIk7iHwHJWwlf2oJUNvHLGjml+gIWXVCTLBlXlgYqUHUVVkIOYxr0FfucIHSZil4vSdVlyBLbPXv4Be/r+/mJrB8r6K2Plm8wNQH7Kt6E/AoGAW7dw3M15w57Zo+jHZ8GOelBPGW+2o7z0jPKtXvns4OjhqslYJmU6yctRWl0Q7zfy+63T85ph8MHeP/9UyuFNkGLiazxKd9yi524xLFKLQXQbs8wkdEMU7qlBU4zHFjuccjrFGygOx0dne/wsSmCpMl2RlBj1vDII6FNU/7nlOkk=\n-----END RSA PRIVATE KEY-----"

/*
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1+tt0BR+tOPPpc/oEgCrYtKvZPwiDgH+h9WSI5zOXZs6wnb4Hj94v//o2wAtYbiKArZDlsJt0G3ylDwVNeBeCUSPjkc0tI80pEB2czfgonE7Y5NogoWE+xQfdRwn2QOw5LD5QPkSKZ+2W8kFzwsnZrEemCXorV4Di2wcpzh7m2zGvc8NUoXbbtvR7RUhTz98NgKNnYCQlS+W2pF/aOiORjq+qqkzbedwLK/WXHGiHrt7Z/2alF+SKwLBMW5Me1TuXghl8wgxShQLu0yN3bwJMNkNMDDE1MfiPeFdDRn12B5ecrn8yZoLjXpBqpiNPX93UUgLToz+BGgypTGT1FcDWwIDAQAB
-----END PUBLIC KEY-----

-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA1+tt0BR+tOPPpc/oEgCrYtKvZPwiDgH+h9WSI5zOXZs6wnb4Hj94v//o2wAtYbiKArZDlsJt0G3ylDwVNeBeCUSPjkc0tI80pEB2czfgonE7Y5NogoWE+xQfdRwn2QOw5LD5QPkSKZ+2W8kFzwsnZrEemCXorV4Di2wcpzh7m2zGvc8NUoXbbtvR7RUhTz98NgKNnYCQlS+W2pF/aOiORjq+qqkzbedwLK/WXHGiHrt7Z/2alF+SKwLBMW5Me1TuXghl8wgxShQLu0yN3bwJMNkNMDDE1MfiPeFdDRn12B5ecrn8yZoLjXpBqpiNPX93UUgLToz+BGgypTGT1FcDWwIDAQABAoIBAHZfWoeeBMz0q80yiv77oPn/mSqa06ysSTd8za56c+R7ip48DNDAaVmRWb5efYK6YecUtz86fmurKzc7LUGpLMSV8sHEpc9rRyfZM1b9RkioHS/9C2mq+3mO0aQpeGsQC/WEVFHbeqqZJadyMJ4Odl5lMemltsb86KKR9a9zVsiguvE6VRHJSEx6/6iXy2JHwS1FSPOT2SzrgI78TSeeROz1CqggJtqZEaAD34SVkzUKst3LvG8Z7UB+mgBzT1V39RwKx70xpBqf9GqhewH5pEOrxUozJ6hrz4sToeDpJ2ubyo7L6tsSFC7pWNciiODCJqtARBUwblOwJx0V1c72lOkCgYEA8yj2k1Zx1QxbVZb0xR56TXRVd3oMNHb9mS0OzFZfL1Bkszlihi8BK2p4gNzQJsRHzmhLND0Uvh7U1v1oF4demiwURgt2dq5bAb1/PtFhKIipdPyiKEPdcxMpEKP6i1UeotdnxomtT31YUS9omVYuiLSUOKuWU5WLZkc9wlRqp/0CgYEA41I7Ah4BeGy5cXk1Haqqo13mL5rUuOkdd5CBR9UHd9xzCxXE2z2L12/J6LcmPxsNheDOFFMLCpSglW16tkFBAOxgjpKidOpI0UG2cIiFRUTpJ8uGygcTfVxqq45jGHs89x54e9AROISweREw0mnI6N+dhMbT9JVy7x1ey9umXDcCgYAvcXeiycQOEIolif2aFFdCk4c1d4+4ENtsLplrjxKlVadAPNsXWUZ+JRj785l9ZuCnyjuaJqzMZ5GZnPnZVWVE6YLPI99qSpyhG0sfg5TUZs3BcKVm+87SbBOgFo6E7we6OBMcbrJtBwTbWkerW2Ba9fjRkdET3+LCAvZu2y+wNQKBgHF7YqPq8Nb6iBVC6iZWRftqa/iF9f4dui0vQarniWPn9LKq+mxsrDwvvX9ktz43tieIk7iHwHJWwlf2oJUNvHLGjml+gIWXVCTLBlXlgYqUHUVVkIOYxr0FfucIHSZil4vSdVlyBLbPXv4Be/r+/mJrB8r6K2Plm8wNQH7Kt6E/AoGAW7dw3M15w57Zo+jHZ8GOelBPGW+2o7z0jPKtXvns4OjhqslYJmU6yctRWl0Q7zfy+63T85ph8MHeP/9UyuFNkGLiazxKd9yi524xLFKLQXQbs8wkdEMU7qlBU4zHFjuccjrFGygOx0dne/wsSmCpMl2RlBj1vDII6FNU/7nlOkk=
-----END RSA PRIVATE KEY-----
*/

func encodeDecode(t *testing.T, senderPublicKey string, senderPrivateKey string, receiverPublicKey string, receiverPrivateKey string, b []byte) {

	cc := GetDefault()
	cc.SenderPublicKey.Value = senderPublicKey
	cc.SenderPrivateKey.Value = senderPrivateKey
	cc.ReceiverPublicKey.Value = receiverPublicKey
	cc.ReceiverPrivateKey.Value = receiverPrivateKey

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
