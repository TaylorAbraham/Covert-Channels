package asymmetricEncryption

import (
	"../../config"
)

type ConfigClient struct {
	SenderPublicKey config.KeyParam
	SenderPrivateKey config.KeyParam
	ReceiverPublicKey config.KeyParam
	ReceiverPrivateKey config.KeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient{
		SenderPublicKey: config.MakeKey(make([]byte, 32), config.Display{Description: "Your Public Key"}),
		SenderPrivateKey: config.MakeKey(make([]byte, 32), config.Display{Description: "Your Private Key"}),
		ReceiverPublicKey: config.MakeKey(make([]byte, 32), config.Display{Description: "Receiver's Public Key"}),
		ReceiverPrivateKey: config.MakeKey(make([]byte, 32), config.Display{Description: "Receiver's Private Key"}),
	}
}

func ToProcessor(cc ConfigClient) (*AsymmetricEncryption, error) {
	return &AsymmetricEncryption{senderPublicKey: cc.SenderPublicKey.Value, senderPrivateKey: cc.SenderPrivateKey.Value, receiverPublicKey: cc.ReceiverPublicKey.Value, receiverPrivateKey: cc.ReceiverPrivateKey.Value}, nil
}
