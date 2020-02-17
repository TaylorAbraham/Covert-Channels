package asymmetricEncryption

import (
	"../../config"
)

type ConfigClient struct {
	SenderPublicKey    config.KeyParam
	SenderPrivateKey   config.KeyParam
	ReceiverPublicKey  config.KeyParam
	ReceiverPrivateKey config.KeyParam
}

func GetDefault() ConfigClient {
	return ConfigClient{
		SenderPublicKey:    config.MakeKey("-----BEGIN RSA PRIVATE KEY-----enter key here-----END RSA PRIVATE KEY-----", config.Display{Description: "Your Public Key", Name: "Sender's Public Key", Group: "Asymmetric Encryption"}),
		SenderPrivateKey:   config.MakeKey("-----BEGIN RSA PRIVATE KEY-----enter key here-----END RSA PRIVATE KEY-----", config.Display{Description: "Your Private Key", Name: "Sender's Private Key", Group: "Asymmetric Encryption"}),
		ReceiverPublicKey:  config.MakeKey("-----BEGIN RSA PRIVATE KEY-----enter key here-----END RSA PRIVATE KEY-----", config.Display{Description: "Receiver's Public Key", Name: "Senders Public Key", Group: "Asymmetric Encryption"}),
		ReceiverPrivateKey: config.MakeKey("-----BEGIN RSA PRIVATE KEY-----enter key here-----END RSA PRIVATE KEY-----", config.Display{Description: "Receiver's Private Key", Name: "Senders Private Key", Group: "Asymmetric Encryption"}),
	}
}

func ToProcessor(cc ConfigClient) (*AsymmetricEncryption, error) {
	return &AsymmetricEncryption{senderPublicKey: []byte(cc.SenderPublicKey.Value), senderPrivateKey: []byte(cc.SenderPrivateKey.Value), receiverPublicKey: []byte(cc.ReceiverPublicKey.Value), receiverPrivateKey: []byte(cc.ReceiverPrivateKey.Value)}, nil
}
