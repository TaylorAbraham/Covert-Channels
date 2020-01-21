package symmetricEncryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"errors"
	"../../config"
)

type ConfigClient struct {
	Algorithm config.SelectParam
	Mode       config.SelectParam
	Key        config.HexKeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient{
		Algorithm: config.MakeSelect("Advanced Encryption Standard (AES)", []string{"Advanced Encryption Standard (AES)", "Data Encryption Standard (DES)", "Triple Data Encryption Standard (3DES)"}, config.Display{Description: "Select an encryption algorithm", Name: "Encryption Algorithm", Group: "Symmetric Encryption"}),
		Mode:       config.MakeSelect("Cipher Block Chaining (CBC)", []string{"Cipher Block Chaining (CBC)", "Cipher Feedback (CFB)", "Counter (CTR)", "Output Feedback (OFB)"}, config.Display{Description: "Select the mode of operation", Name: "Mode of Operation", Group: "Symmetric Encryption"}),
		// AES-128 = key size 32 characters long, AES-192 = key size 48 
		//characters long, and AES-256 = key size 64 characters long
		// DES key size 16 characters long, 3DES key size 48 characters
		// long (i.e. 3*16)
		Key: config.MakeHexKey(make([]byte, 32), []int{8, 16, 24, 32}, config.Display{Description: "The shared secret key used for Advanced Encryption Standard (AES) must be 32, 48 or 64 characters in length, for Data Encryption Standard (DES) must be 16 characters in length, and for Triple Data Encryption Standard (3DES) must be 48 characters in length", Name: "Shared Secret Key", Group: "Symmetric Encryption"})}
}

func ToProcessor(cc ConfigClient) (*SymmetricEncryption, error) {
	var (
		blockSize int
		block     cipher.Block
		err       error
	)

	// based on the users choice of symmetric algorithm create a cipher
	switch cc.Algorithm.Value {
	case "Advanced Encryption Standard (AES)":
		block, err = aes.NewCipher(cc.Key.Value)
		blockSize = 16
	case "Data Encryption Standard (DES)":
		block, err = des.NewCipher(cc.Key.Value)
		blockSize = 8
	case "Triple Data Encryption Standard (3DES)":
		block, err = des.NewTripleDESCipher(cc.Key.Value)
		blockSize = 8
	default:
		return nil, errors.New("Undefined algorithm selected")
	}

	if err != nil {
		return nil, err
	}

	return &SymmetricEncryption{algorithm: cc.Algorithm.Value, mode: cc.Mode.Value, key: cc.Key.Value, block: block, blockSize: blockSize}, nil
}
