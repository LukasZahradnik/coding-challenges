package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
)

type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

type RSASigner struct {
	PrivateKey *rsa.PrivateKey
}

func NewRSASigner(privateKey []byte) (*RSASigner, error) {
	key, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &RSASigner{PrivateKey: key}, nil
}

func (s *RSASigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	hashed := sha256.Sum256(dataToBeSigned)
	return rsa.SignPKCS1v15(rand.Reader, s.PrivateKey, crypto.SHA256, hashed[:])
}

type ECDSASigner struct {
	PrivateKey *ecdsa.PrivateKey
}

func NewECDSASigner(privateKey []byte) (*ECDSASigner, error) {
	key, err := x509.ParseECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &ECDSASigner{PrivateKey: key}, nil
}

func (s *ECDSASigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	hashed := sha256.Sum256(dataToBeSigned)
	return ecdsa.SignASN1(rand.Reader, s.PrivateKey, hashed[:])
}
