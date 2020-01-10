package advancedEncryptionStandard

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

const padIndex = 7;

type AdvancedEncryptionStandard struct {
	mode string
	key []byte
}

func(c *AdvancedEncryptionStandard) Process(data []byte) ([]byte, error) {
	data = Pad(data)

	fmt.Print("Checkpoint1")
	block, err := aes.NewCipher(c.key)
	if err != nil {
		fmt.Print("Checkpoint2")
		return nil, err
		fmt.Print("Checkpoint3")
	}

	fmt.Print("Checkpoint4")

	selectedMode := c.mode
	switch selectedMode {
	case "Galois Counter Mode (GCM)":
		fmt.Print("GCM")
	case "Cipher Block Chaining (CBC)":
		fmt.Print("CBC")
	case "Cipher Feedback (CFB)":
		fmt.Print("CFB")
	case "Counter (CTR)":
		fmt.Print("CTR")
	default: 
		return nil, errors.New("Undefined mode selected")
	}
	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], data)

	return cipherText[:], nil
}

func Pad(data []byte) ([]byte) {
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

func UnPad(data []byte) ([]byte) {
	lenOfData := data[padIndex]
	data = data[padIndex+1:]
	return data[:lenOfData]
}

func(c *AdvancedEncryptionStandard) Unprocess(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)

	data = UnPad(data)

	return data[:], nil
} 