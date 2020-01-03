package caesar

import (
	"../../config"
)

type ConfigClient struct {
	Shift config.I8Param
}

func GetDefault() ConfigClient {
	return ConfigClient{
		// Julius Caesar crossed the Rubicon in 49 AD
		Shift: config.MakeI8(49, [2]int8{-128, 127}, config.Display{Description: "The shift for the Caesar cypher."})}
}

func ToProcessor(cc ConfigClient) (*Caesar, error) {
	return &Caesar{Shift: cc.Shift.Value}, nil
}
