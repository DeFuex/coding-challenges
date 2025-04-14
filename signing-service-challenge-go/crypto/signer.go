package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// Signer defines a contract for different types of signing implementations.
type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

// RSASigner implements the Signer interface for RSA signatures
type RSASigner struct {
	privateKey *rsa.PrivateKey
}

// NewRSASigner creates a new RSA signer
func NewRSASigner(privateKey *rsa.PrivateKey) *RSASigner {
	return &RSASigner{
		privateKey: privateKey,
	}
}

// Sign signs the provided data using RSA
func (signer *RSASigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	hashed := sha256.Sum256(dataToBeSigned)
	return rsa.SignPKCS1v15(rand.Reader, signer.privateKey, crypto.SHA256, hashed[:])
}

// ECCSigner implements the Signer interface for ECDSA signatures
type ECCSigner struct {
	privateKey *ecdsa.PrivateKey
}

// NewECCSigner creates a new ECC signer
func NewECCSigner(privateKey *ecdsa.PrivateKey) *ECCSigner {
	return &ECCSigner{
		privateKey: privateKey,
	}
}

// Sign signs the provided data using ECDSA
func (signer *ECCSigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	hashed := sha256.Sum256(dataToBeSigned)
	r, sig, err := ecdsa.Sign(rand.Reader, signer.privateKey, hashed[:])
	if err != nil {
		return nil, err
	}

	// Concatenate r and s values
	signature := make([]byte, 96) // 48 bytes for r, 48 bytes for s
	r.FillBytes(signature[:48])
	sig.FillBytes(signature[48:])

	return signature, nil
}

// TODO: implement RSA and ECDSA signing ...
