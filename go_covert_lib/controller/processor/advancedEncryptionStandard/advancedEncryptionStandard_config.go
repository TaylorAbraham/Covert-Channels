package advancedEncryptionStandard

import (
	"../../config"
)

type ConfigClient struct {
	Key config.HexKeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient {
		Key: config.MakeHexKey(make([]byte, 32), []int{16, 24, 32}, config.Display{Description: "The shared secret key used for symmtric encryption."})}
} 

func ToProcessor(aes ConfigClient) (*AdvancedEncryptionStandard, error) {
	return &AdvancedEncryptionStandard{key: aes.Key.Value}, nil
}