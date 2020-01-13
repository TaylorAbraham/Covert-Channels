package asymmetricEncryption

type AsymmetricEncryption struct {
	senderPublicKey []byte
	senderPrivateKey []byte
	receiverPublicKey []byte
	receiverPrivateKey []byte
}

func (c *AsymmetricEncryption) Process(data []byte) ([]byte, error) {
	return nil, nil
}

func (c *AsymmetricEncryption) Unprocess(data []byte) ([]byte, error) {
	return nil, nil
}