package checksum

import (
	"../../config"
	"errors"
	"hash/crc32"
)

type ConfigClient struct {
	Polynomial config.SelectParam
}

func GetDefault() ConfigClient {
	return ConfigClient{
		Polynomial: config.MakeSelect("IEEE", []string{"IEEE", "Castagnoli", "Koopman"}, config.Display{Description: "The predefined polynomial to use for the crc32 checksum"}),
	}
}

func ToProcessor(cc ConfigClient) (*Checksum, error) {
	var poly uint32
	switch cc.Polynomial.Value {
	case "IEEE":
		poly = crc32.IEEE
	case "Castagnoli":
		poly = crc32.Castagnoli
	case "Koopman":
		poly = crc32.Koopman
	default:
		return nil, errors.New("Invalid polynomial")
	}
	return &Checksum{table: crc32.MakeTable(poly)}, nil
}
