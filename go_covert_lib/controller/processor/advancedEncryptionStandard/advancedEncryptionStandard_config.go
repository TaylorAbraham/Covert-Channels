package advancedEncryptionStandard

import (
	"../../config"
)

type ConfigClient struct {
	Algorithms config.SelectParam
	Mode config.SelectParam
	Key config.HexKeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient {
		Algorithms: config.MakeSelect("Advanced Encryption Standard (AES)", []string{"Advanced Encryption Standard (AES)", "Data Encryption Standard (DES)", "Triple Data Encryption Standard (3DES)"}, config.Display{Description: "Select an encryption algorithm"}),
		Mode: config.MakeSelect("Galois Counter Mode (GCM)", []string{"Galois Counter Mode (GCM)", "Cipher Block Chaining (CBC)", "Cipher Feedback (CFB)", "Counter (CTR)"}, config.Display{Description: "Select the mode of operation"}),
		// AES-128 = key size 32 characters long, AES-192 = key size 48 characters long, and AES-256 = key size 64 characters long
		Key: config.MakeHexKey(make([]byte, 32), []int{8, 16, 24, 32}, config.Display{Description: "The shared secret key used for symmtric encryption must be 32, 48 or 64 characters in length"})}
} 

func ToProcessor(aes ConfigClient) (*AdvancedEncryptionStandard, error) {
	return &AdvancedEncryptionStandard{algorithm: aes.Algorithms.Value, mode: aes.Mode.Value, key: aes.Key.Value}, nil
}