package advancedEncryptionStandard

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type AdvancedEncryptionStandard struct {
	key []byte
}

func(c *AdvancedEncryptionStandard) Process(data []byte) ([]byte, error) {
	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("The message is not a multiple of the block size")
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], data)

	return cipherText[:], nil
}

func(c *AdvancedEncryptionStandard) Unprocess(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, errors.New("The message is too short for given key")
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	return data[:], nil
} 