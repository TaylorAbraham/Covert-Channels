package symmetricEncryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"errors"
)

// padding is done by having 7 zero's at the start followed by the length of 
// the original message, the contents of the message, followed by a series of
// zero's to fill the block size requirements of the symmetric encryption 
// algorithm 
const padIndex = 7

type SymmetricEncryption struct {
	algorithm string
	mode      string
	key       []byte
}

func (c *SymmetricEncryption) Process(data []byte) ([]byte, error) {
	data = Pad(data)
	algorithm := c.algorithm

	var (
		blockSize int
		block     cipher.Block
		err       error
	)

	// based on the users choice of symmetric algorithm create a cipher
	switch algorithm {
	case "Advanced Encryption Standard (AES)":
		block, err = aes.NewCipher(c.key)
		blockSize = 16
	case "Data Encryption Standard (DES)":
		block, err = des.NewCipher(c.key)
		blockSize = 8
	case "Triple Data Encryption Standard (3DES)":
		block, err = des.NewTripleDESCipher(c.key)
		blockSize = 8
	default:
		return nil, errors.New("Undefined algorithm selected")
	}

	if err != nil {
		return nil, err
	}

	// based on the users choice of the mode of operation encrypt in that mode
	var cipherText []byte
	selectedMode := c.mode
	switch selectedMode {
	case "Galois Counter Mode (GCM)":
		cipherText = GCMEncrypter(block, data)
		if cipherText == nil {
			return nil, errors.New("Unable to encrypt in Galosis Counter Mode (GCM)")
		}
	case "Cipher Block Chaining (CBC)":
		cipherText = CBCEncrypter(block, data, blockSize)
	case "Cipher Feedback (CFB)":
		cipherText = CFBEncrypter(block, data, blockSize)
	case "Counter (CTR)":
		cipherText = CTREncrypter(block, data, blockSize)
	default:
		return nil, errors.New("Undefined mode selected")
	}

	return cipherText[:], nil
}

func GCMEncrypter(block cipher.Block, data []byte) []byte {
	nonce := make([]byte, 12)

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	cipherText := aesgcm.Seal(nil, nonce, data, nil)

	return cipherText[:]
}

func CBCEncrypter(block cipher.Block, data []byte, blockSize int) []byte {
	cipherText := make([]byte, blockSize+len(data))
	iv := cipherText[:blockSize]

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[blockSize:], data)

	return cipherText[:]
}

func CFBEncrypter(block cipher.Block, data []byte, blockSize int) []byte {
	cipherText := make([]byte, blockSize+len(data))
	iv := cipherText[:blockSize]

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[blockSize:], data)

	return cipherText[:]
}

func CTREncrypter(block cipher.Block, data []byte, blockSize int) []byte {
	cipherText := make([]byte, blockSize+len(data))
	iv := cipherText[:blockSize]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText[blockSize:], data)

	return cipherText[:]
}

func Pad(data []byte) []byte {
	// padding is done by having 7 zero's at the start followed by the length 
	// of the original message, the contents of the message, followed by a 
	// series of zero's to fill the block size requirements of the symmetric  
	// encryption algorithm 
	lenOfData := len(data)
	data = append(data, 0)
	copy(data[1:], data)
	data[0] = byte(lenOfData)

	for i := 0; i < padIndex; i++ {
		data = append(data, 0)
		copy(data[1:], data)
		data[0] = 0
	}

	for len(data)%aes.BlockSize != 0 {
		data = append(data, 0)
	}

	return data[:]
}

func UnPad(data []byte) []byte {
	lenOfData := data[padIndex]
	data = data[padIndex+1:]
	return data[:lenOfData]
}

func (c *SymmetricEncryption) Unprocess(data []byte) ([]byte, error) {
	algorithm := c.algorithm

	var (
		blockSize int
		block     cipher.Block
		err       error
	)

	// based on the users choice of symmetric algorithm create a cipher
	switch algorithm {
	case "Advanced Encryption Standard (AES)":
		block, err = aes.NewCipher(c.key)
		blockSize = 16
	case "Data Encryption Standard (DES)":
		block, err = des.NewCipher(c.key)
		blockSize = 8
	case "Triple Data Encryption Standard (3DES)":
		block, err = des.NewTripleDESCipher(c.key)
		blockSize = 8
	default:
		return nil, errors.New("Undefined algorithm selected")
	}

	if err != nil {
		return nil, err
	}

	// based on the users choice of the mode of operation encrypt in that mode
	selectedMode := c.mode
	switch selectedMode {
	case "Galois Counter Mode (GCM)":
		data = GCMDecrypter(block, data)
		if data == nil {
			return nil, errors.New("Unable to decrypt in Galosis Counter Mode (GCM)")
		}
	case "Cipher Block Chaining (CBC)":
		data = CBCDecrypter(block, data, blockSize)
	case "Cipher Feedback (CFB)":
		data = CFBDecrypter(block, data, blockSize)
	case "Counter (CTR)":
		data = CTRDecrypter(block, data, blockSize)
	default:
		return nil, errors.New("Undefined mode selected")
	}

	data = UnPad(data)

	return data[:], nil
}

func GCMDecrypter(block cipher.Block, data []byte) []byte {
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	nonce := make([]byte, 12)
	plaintext, err := aesgcm.Open(nil, nonce, data, nil)

	return plaintext[:]
}

func CBCDecrypter(block cipher.Block, data []byte, blockSize int) []byte {
	iv := data[:blockSize]
	data = data[blockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	return data[:]
}

func CFBDecrypter(block cipher.Block, data []byte, blockSize int) []byte {
	iv := data[:blockSize]
	data = data[blockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return data[:]
}

func CTRDecrypter(block cipher.Block, data []byte, blockSize int) []byte {
	plaintext := make([]byte, len(data))
	iv := data[:blockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, data[blockSize:])

	return plaintext[:]
}
