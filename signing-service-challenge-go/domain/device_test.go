package domain

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/crypto"
)

func TestNewSignatureDevice(t *testing.T) {
	// Setup
	id := "test-device"
	algorithm := ECC
	label := "Test Device"
	signer, privateKey, publicKey, err := crypto.NewECCSigner()
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Test
	device := NewSignatureDevice(id, algorithm, label, signer, privateKey, publicKey)

	// Verify
	if device.ID != id {
		t.Errorf("Expected device ID %s, got %s", id, device.ID)
	}
	if device.Algorithm != algorithm {
		t.Errorf("Expected algorithm %s, got %s", algorithm, device.Algorithm)
	}
	if device.Label != label {
		t.Errorf("Expected label %s, got %s", label, device.Label)
	}
	if device.SignatureCounter != 0 {
		t.Errorf("Expected signature counter 0, got %d", device.SignatureCounter)
	}
	expectedInitialSignature := base64.StdEncoding.EncodeToString([]byte(id))
	if device.LastSignature != expectedInitialSignature {
		t.Errorf("Expected last signature %s, got %s", expectedInitialSignature, device.LastSignature)
	}
}

func TestSignTransaction(t *testing.T) {
	// Setup
	id := "test-device"
	algorithm := ECC
	label := "Test Device"
	signer, privateKey, publicKey, err := crypto.NewECCSigner()
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}
	device := NewSignatureDevice(id, algorithm, label, signer, privateKey, publicKey)
	data := "test-data"

	// Test first signature
	sig1, err := device.SignTransaction(data)
	if err != nil {
		t.Fatalf("Failed to sign first transaction: %v", err)
	}

	// Verify first signature
	if device.SignatureCounter != 1 {
		t.Errorf("Expected signature counter 1, got %d", device.SignatureCounter)
	}
	if !strings.Contains(sig1["signed_data"], data) {
		t.Errorf("Signed data does not contain original data: %s", sig1["signed_data"])
	}
	expectedFormat := "0_test-data_"
	if !strings.HasPrefix(sig1["signed_data"], expectedFormat) {
		t.Errorf("Expected signed data to start with %s, got %s", expectedFormat, sig1["signed_data"])
	}

	// Test second signature
	sig2, err := device.SignTransaction(data)
	if err != nil {
		t.Fatalf("Failed to sign second transaction: %v", err)
	}

	// Verify second signature
	if device.SignatureCounter != 2 {
		t.Errorf("Expected signature counter 2, got %d", device.SignatureCounter)
	}
	if sig2["signed_data"] == sig1["signed_data"] {
		t.Error("Second signature should be different from first signature")
	}
	if !strings.Contains(sig2["signed_data"], sig1["signature"]) {
		t.Error("Second signature should contain first signature")
	}
}

func TestConcurrentSignTransaction(t *testing.T) {
	// Setup
	id := "test-device"
	algorithm := ECC
	label := "Test Device"
	signer, privateKey, publicKey, err := crypto.NewECCSigner()
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}
	device := NewSignatureDevice(id, algorithm, label, signer, privateKey, publicKey)

	// Test concurrent signatures
	numGoroutines := 10
	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			data := "test-data"
			_, err := device.SignTransaction(data)
			if err != nil {
				t.Errorf("Failed to sign transaction in goroutine %d: %v", index, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state
	if device.SignatureCounter != numGoroutines {
		t.Errorf("Expected signature counter %d, got %d", numGoroutines, device.SignatureCounter)
	}
}
