package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
)

// Signer defines a contract for different types of signing implementations.
type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

// RSASigner implements the Signer interface for RSA signatures
type RSASigner struct {
	privateKey *rsa.PrivateKey
}

// NewRSASigner creates a new RSA signer with a new key pair
func NewRSASigner() (*RSASigner, []byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &RSASigner{
		privateKey: privateKey,
	}, privateKeyPEM, publicKeyPEM, nil
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

// NewECCSigner creates a new ECC signer with a new key pair
func NewECCSigner() (*ECCSigner, []byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	// Encode private key
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, nil, err
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &ECCSigner{
		privateKey: privateKey,
	}, privateKeyPEM, publicKeyPEM, nil
}

// Sign signs the provided data using ECDSA
func (signer *ECCSigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	hashed := sha256.Sum256(dataToBeSigned)
	r, s, err := ecdsa.Sign(rand.Reader, signer.privateKey, hashed[:])
	if err != nil {
		return nil, err
	}

	// Concatenate r and s values
	signature := make([]byte, 96) // 48 bytes for r, 48 bytes for s
	r.FillBytes(signature[:48])
	s.FillBytes(signature[48:])

	return signature, nil
}

// TODO: implement RSA and ECDSA signing ...
