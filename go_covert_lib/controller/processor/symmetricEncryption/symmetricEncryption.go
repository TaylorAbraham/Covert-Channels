package symmetricEncryption

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"encoding/binary"
)

// padding is done by having 7 zero's at the start followed by the length of 
// the original message, the contents of the message, followed by a series of
// zero's to fill the block size requirements of the symmetric encryption 
// algorithm 
const padAmount = 8

type SymmetricEncryption struct {
	algorithm string
	mode      string
	key       []byte
	block	cipher.Block
	blockSize int
}

func (c *SymmetricEncryption) Process(data []byte) ([]byte, error) {
	data = Pad(data)
	// based on the users choice of the mode of operation encrypt in that mode
	var cipherText []byte
	switch c.mode {
	case "Cipher Block Chaining (CBC)":
		cipherText = CBCEncrypter(c.block, data, c.blockSize)
	case "Cipher Feedback (CFB)":
		cipherText = CFBEncrypter(c.block, data, c.blockSize)
	case "Counter (CTR)":
		cipherText = CTREncrypter(c.block, data, c.blockSize)
	case "Output Feedback (OFB)":
		cipherText = OFBEncrypter(c.block, data, c.blockSize)
	default:
		return nil, errors.New("Undefined mode selected")
	}

	return cipherText, nil
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

func OFBEncrypter(block cipher.Block, data []byte, blockSize int) []byte {
	cipherText := make([]byte, blockSize+len(data))
	iv := cipherText[:blockSize]

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(cipherText[blockSize:], data)

	return cipherText[:]
}

func Pad(data []byte) []byte {
	// padding is done by having 7 zero's at the start followed by the length 
	// of the original message, the contents of the message, followed by a 
	// series of zero's to fill the block size requirements of the symmetric  
	// encryption algorithm 

	b := make([]byte, padAmount)
	b = append(b, data...)
	binary.BigEndian.PutUint64(b[:padAmount], uint64(len(data)))

	for len(b)%aes.BlockSize != 0 {
		b = append(b, 0)
	}
	return b
}

func UnPad(data []byte) ([]byte, error) {
	lenOfData := binary.BigEndian.Uint64(data[:padAmount])
	if lenOfData < uint64(len(data) - 8) {
		return nil, errors.New("Failed to unprocess data, as data was not processed correctly")
	}
	data = data[padAmount:]
	return data[:lenOfData], nil
}

func (c *SymmetricEncryption) Unprocess(data []byte) ([]byte, error) {
	if len(data)%c.blockSize != 0 {
		return nil, errors.New("Unable to decrypt, wrong block size")
	}
	
	// based on the users choice of the mode of operation encrypt in that mode
	switch c.mode {
	case "Cipher Block Chaining (CBC)":
		data = CBCDecrypter(c.block, data, c.blockSize)
	case "Cipher Feedback (CFB)":
		data = CFBDecrypter(c.block, data, c.blockSize)
	case "Counter (CTR)":
		data = CTRDecrypter(c.block, data, c.blockSize)
	case "Output Feedback (OFB)":
		data = OFBDecrypter(c.block, data, c.blockSize)
	default:
		return nil, errors.New("Undefined mode selected")
	}

	data, _ = UnPad(data)

	return data, nil
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

func OFBDecrypter(block cipher.Block, data []byte, blockSize int) []byte {
	plaintext := make([]byte, len(data))
	iv := data[:blockSize]
	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(plaintext, data[blockSize:])

	return plaintext[:]
}
