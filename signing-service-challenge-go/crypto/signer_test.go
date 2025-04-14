package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestNewRSASigner(t *testing.T) {
	// Test signer creation
	signer, privateKeyPEM, publicKeyPEM, err := NewRSASigner()
	if err != nil {
		t.Fatalf("NewRSASigner() error = %v", err)
	}
	if signer == nil {
		t.Fatal("NewRSASigner() returned nil signer")
	}

	// Test private key format
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		t.Fatal("Failed to decode private key PEM")
	}
	if block.Type != "RSA PRIVATE KEY" {
		t.Errorf("Private key type = %v, want RSA PRIVATE KEY", block.Type)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	if privateKey.Size() != 256 { // 2048 bits = 256 bytes
		t.Errorf("Private key size = %v bytes, want 256 bytes", privateKey.Size())
	}

	// Test public key format
	block, _ = pem.Decode(publicKeyPEM)
	if block == nil {
		t.Fatal("Failed to decode public key PEM")
	}
	if block.Type != "PUBLIC KEY" {
		t.Errorf("Public key type = %v, want PUBLIC KEY", block.Type)
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}
	if _, ok := publicKey.(*rsa.PublicKey); !ok {
		t.Error("Public key is not an RSA public key")
	}
}

func TestNewECCSigner(t *testing.T) {
	// Test signer creation
	signer, privateKeyPEM, publicKeyPEM, err := NewECCSigner()
	if err != nil {
		t.Fatalf("NewECCSigner() error = %v", err)
	}
	if signer == nil {
		t.Fatal("NewECCSigner() returned nil signer")
	}

	// Test private key format
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		t.Fatal("Failed to decode private key PEM")
	}
	if block.Type != "EC PRIVATE KEY" {
		t.Errorf("Private key type = %v, want EC PRIVATE KEY", block.Type)
	}
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	if privateKey.Curve.Params().Name != "P-256" {
		t.Errorf("Curve = %v, want P-256", privateKey.Curve.Params().Name)
	}

	// Test public key format
	block, _ = pem.Decode(publicKeyPEM)
	if block == nil {
		t.Fatal("Failed to decode public key PEM")
	}
	if block.Type != "PUBLIC KEY" {
		t.Errorf("Public key type = %v, want PUBLIC KEY", block.Type)
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}
	if _, ok := publicKey.(*ecdsa.PublicKey); !ok {
		t.Error("Public key is not an ECDSA public key")
	}
}

func TestRSASignerSign(t *testing.T) {
	signer, _, _, err := NewRSASigner()
	if err != nil {
		t.Fatalf("Failed to create RSA signer: %v", err)
	}

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "Sign simple data",
			data:    []byte("test data"),
			wantErr: false,
		},
		{
			name:    "Sign empty data",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "Sign large data",
			data:    bytes.Repeat([]byte("a"), 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig1, err := signer.Sign(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// RSA signatures should be deterministic
				sig2, err := signer.Sign(tt.data)
				if err != nil {
					t.Errorf("Second Sign() failed: %v", err)
					return
				}
				if !bytes.Equal(sig1, sig2) {
					t.Error("RSA signatures for same data are not identical")
				}
			}
		})
	}
}

func TestECCSignerSign(t *testing.T) {
	signer, _, _, err := NewECCSigner()
	if err != nil {
		t.Fatalf("Failed to create ECC signer: %v", err)
	}

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "Sign simple data",
			data:    []byte("test data"),
			wantErr: false,
		},
		{
			name:    "Sign empty data",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "Sign large data",
			data:    bytes.Repeat([]byte("a"), 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig1, err := signer.Sign(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(sig1) != 96 { // 48 bytes for r + 48 bytes for s
					t.Errorf("Signature length = %v, want 96", len(sig1))
				}

				// ECDSA signatures should be non-deterministic
				sig2, err := signer.Sign(tt.data)
				if err != nil {
					t.Errorf("Second Sign() failed: %v", err)
					return
				}
				if bytes.Equal(sig1, sig2) {
					t.Error("ECDSA signatures for same data are identical (should be different)")
				}
			}
		})
	}
}
