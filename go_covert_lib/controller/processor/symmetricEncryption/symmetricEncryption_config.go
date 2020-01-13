package symmetricEncryption

import (
	"../../config"
)

type ConfigClient struct {
	Algorithm config.SelectParam
	Mode       config.SelectParam
	Key        config.HexKeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient{
		Algorithm: config.MakeSelect("Advanced Encryption Standard (AES)", []string{"Advanced Encryption Standard (AES)", "Data Encryption Standard (DES)", "Triple Data Encryption Standard (3DES)"}, config.Display{Description: "Select an encryption algorithm"}),
		Mode:       config.MakeSelect("Galois Counter Mode (GCM)", []string{"Galois Counter Mode (GCM)", "Cipher Block Chaining (CBC)", "Cipher Feedback (CFB)", "Counter (CTR)"}, config.Display{Description: "Select the mode of operation"}),
		// AES-128 = key size 32 characters long, AES-192 = key size 48 
		//characters long, and AES-256 = key size 64 characters long
		// DES key size 16 characters long, 3DES key size 48 characters
		// long (i.e. 3*16)
		// NOTE: key byte is 2 characters hence the options are 8, 16, 24, or 32
		Key: config.MakeHexKey(make([]byte, 32), []int{8, 16, 24, 32}, config.Display{Description: "The shared secret key used for Advanced Encryption Standard (AES) must be 32, 48 or 64 characters in length, for Data Encryption Standard (DES) must be 16 characters in length, and for Triple Data Encryption Standard (3DES) must be 48 characters in length"})}
}

func ToProcessor(aes ConfigClient) (*SymmetricEncryption, error) {
	return &SymmetricEncryption{algorithm: aes.Algorithm.Value, mode: aes.Mode.Value, key: aes.Key.Value}, nil
}
