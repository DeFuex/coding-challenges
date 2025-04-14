package domain

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/crypto"
)

// SignatureAlgorithm represents the supported signature algorithms
type SignatureAlgorithm string

const (
	// ECC represents the ECDSA signature algorithm
	ECC SignatureAlgorithm = "ECC"
	// RSA represents the RSA signature algorithm
	RSA SignatureAlgorithm = "RSA"
)

// SignatureDevice represents a device that can sign transaction data
type SignatureDevice struct {
	ID               string             `json:"id"`
	Label            string             `json:"label"`
	Algorithm        SignatureAlgorithm `json:"algorithm"`
	SignatureCounter int                `json:"signature_counter"`
	LastSignature    string             `json:"last_signature"`
	PrivateKey       []byte             `json:"-"`
	PublicKey        []byte             `json:"public_key"`
	Signer           crypto.Signer      `json:"-"`
	mu               sync.Mutex         `json:"-"`
}

// NewSignatureDevice creates a new signature device with the given parameters
func NewSignatureDevice(id string, algorithm SignatureAlgorithm, label string, signer crypto.Signer, privateKey, publicKey []byte) *SignatureDevice {
	return &SignatureDevice{
		ID:               id,
		Label:            label,
		Algorithm:        algorithm,
		SignatureCounter: 0,
		LastSignature:    base64.StdEncoding.EncodeToString([]byte(id)), // Initial last signature is base64 encoded device ID
		PrivateKey:       privateKey,
		PublicKey:        publicKey,
		Signer:           signer,
	}
}

// SignTransaction signs the provided data and returns the signature response
func (d *SignatureDevice) SignTransaction(data string) (map[string]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create the secured data string
	securedData := fmt.Sprintf("%d_%s_%s", d.SignatureCounter, data, d.LastSignature)

	// Sign the secured data
	signature, err := d.Signer.Sign([]byte(securedData))
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Encode the signature to base64
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	// Update the device state
	d.SignatureCounter++
	d.LastSignature = signatureBase64

	// Return the signature response
	return map[string]string{
		"signature":   signatureBase64,
		"signed_data": securedData,
	}, nil
}
