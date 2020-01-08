package advancedEncryptionStandard

import (
	"math"
	"encoding/binary"
	"../../config"
)

type ConfigClient struct {
	Key config.U32Param
}

func GetDefault() ConfigClient {
	return ConfigClient {
		Key: config.MakeU32(0, [2]uint64{0, math.MaxUint32}, config.Display{Description: "The shared secret key used for symmtric encryption."})}
} 

func ToProcessor(aes ConfigClient) (*AdvancedEncryptionStandard, error) {
	byteArray := make([]byte, 32)
	binary.LittleEndian.PutUint64(byteArray, aes.Key.Value)
	return &AdvancedEncryptionStandard{key: byteArray}, nil
}