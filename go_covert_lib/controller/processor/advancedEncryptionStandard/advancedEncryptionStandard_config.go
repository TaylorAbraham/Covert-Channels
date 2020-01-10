package advancedEncryptionStandard

import (
	"../../config"
)

type ConfigClient struct {
	Mode config.SelectParam
	Key config.HexKeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient {
		Mode: config.MakeSelect("Galois Counter Mode (GCM)", []string{"Galois Counter Mode (GCM)", "Cipher Block Chaining (CBC)", "Cipher Feedback (CFB)", "Counter (CTR)"}, config.Display{Description: "Select the mode of operation."}),
		Key: config.MakeHexKey(make([]byte, 32), []int{16, 24, 32}, config.Display{Description: "The shared secret key used for symmtric encryption."})}
} 

func ToProcessor(aes ConfigClient) (*AdvancedEncryptionStandard, error) {
	return &AdvancedEncryptionStandard{mode: aes.Mode.Value, key: aes.Key.Value}, nil
}