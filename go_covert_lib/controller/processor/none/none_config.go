package none

import (
	"../../config"
)

type ConfigClient struct {
	Exactu64Test config.ExactU64Param
	Keytest      config.HexKeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient{
		Exactu64Test: config.MakeExactU64(12345, config.Display{Description: "Test for exact u64 type"}),
		Keytest:      config.MakeHexKey([]byte{1, 2, 3, 4}, []int{2, 4, 8}, config.Display{Description: "Test for key type"}),
	}
}

func ToProcessor(cc ConfigClient) (*None, error) {
	return &None{}, nil
}
