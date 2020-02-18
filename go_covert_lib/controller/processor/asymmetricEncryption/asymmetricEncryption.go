package asymmetricEncryption

import (
	"crypto/sha512"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type AsymmetricEncryption struct {
	senderPublicKey []byte
	senderPrivateKey []byte
	receiverPublicKey []byte
	receiverPrivateKey []byte
}

func (c *AsymmetricEncryption) Process(data []byte) ([]byte, error) {
	hash := sha512.New()
	rsaPublicKey, err := BytesToPublicKey(c.receiverPublicKey)
	if err != nil {
		return nil, err
	}

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, rsaPublicKey, data, nil)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

func (c *AsymmetricEncryption) Unprocess(data []byte) ([]byte, error) {
	hash := sha512.New()

	rsaPrivateKey, err := BytesToPrivateKey(c.receiverPrivateKey)
	if err != nil {
		return nil, err
	}

	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, rsaPrivateKey, data, nil)

	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func PrivateKeyToBytes(privateKey *rsa.PrivateKey) []byte {
	privateKeyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	return privateKeyBytes
}

func PublicKeyToBytes(publicKey *rsa.PublicKey) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	publicKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubBytes,
	})

	return publicKeyBytes, nil
}

func BytesToPrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKey)

	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("Failed to decode PEM block containing private key")
	}
	encrypted := x509.IsEncryptedPEMBlock(block)

	b := block.Bytes
	var err error
	if encrypted {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, err
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func BytesToPublicKey(publicKey []byte) (*rsa.PublicKey, error) {
	var err error
	block, _ := pem.Decode(publicKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("Failed to decode PEM block containing public key")
	}
	encrypted := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes
	if encrypted {
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			return nil, err
		}
	}
	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, err
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("Not type *rsa.PublicKey")
	}
	return key, nil
}
