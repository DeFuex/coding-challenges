package persistence

import (
	"fmt"
	"sync"
	"testing"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
)

func TestNewInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()
	if store == nil {
		t.Fatal("Expected non-nil store")
	}
	if store.devices == nil {
		t.Fatal("Expected devices map to be initialized")
	}
}

func TestCreate(t *testing.T) {
	store := NewInMemoryStore()

	tests := []struct {
		name      string
		id        string
		algorithm domain.SignatureAlgorithm
		label     string
		wantErr   error
	}{
		{
			name:      "Create ECC Device",
			id:        "test-ecc",
			algorithm: domain.ECC,
			label:     "Test ECC Device",
			wantErr:   nil,
		},
		{
			name:      "Create RSA Device",
			id:        "test-rsa",
			algorithm: domain.RSA,
			label:     "Test RSA Device",
			wantErr:   nil,
		},
		{
			name:      "Create Device with Invalid Algorithm",
			id:        "test-invalid",
			algorithm: "INVALID",
			label:     "Test Invalid Device",
			wantErr:   domain.ErrInvalidAlgorithm,
		},
		{
			name:      "Create Duplicate Device",
			id:        "test-ecc", // Same as first test
			algorithm: domain.ECC,
			label:     "Duplicate Device",
			wantErr:   domain.ErrDeviceAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device, err := store.Create(tt.id, tt.algorithm, tt.label)
			if err != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if device.ID != tt.id {
					t.Errorf("Create() device.ID = %v, want %v", device.ID, tt.id)
				}
				if device.Algorithm != tt.algorithm {
					t.Errorf("Create() device.Algorithm = %v, want %v", device.Algorithm, tt.algorithm)
				}
				if device.Label != tt.label {
					t.Errorf("Create() device.Label = %v, want %v", device.Label, tt.label)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	store := NewInMemoryStore()

	// Create a test device
	device, err := store.Create("test-device", domain.ECC, "Test Device")
	if err != nil {
		t.Fatalf("Failed to create test device: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		want    *domain.SignatureDevice
		wantErr error
	}{
		{
			name:    "Get Existing Device",
			id:      device.ID,
			want:    device,
			wantErr: nil,
		},
		{
			name:    "Get Non-existent Device",
			id:      "non-existent",
			want:    nil,
			wantErr: domain.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Get(tt.id)
			if err != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got.ID != tt.want.ID {
				t.Errorf("Get() = %v, want %v", got.ID, tt.want.ID)
			}
		})
	}
}

func TestList(t *testing.T) {
	store := NewInMemoryStore()

	// Test empty list
	devices, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("List() = %v, want empty list", devices)
	}

	// Create test devices
	testDevices := []struct {
		id        string
		algorithm domain.SignatureAlgorithm
		label     string
	}{
		{"device1", domain.ECC, "Device 1"},
		{"device2", domain.RSA, "Device 2"},
		{"device3", domain.ECC, "Device 3"},
	}

	for _, d := range testDevices {
		_, err := store.Create(d.id, d.algorithm, d.label)
		if err != nil {
			t.Fatalf("Failed to create test device: %v", err)
		}
	}

	// Test list with devices
	devices, err = store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(devices) != len(testDevices) {
		t.Errorf("List() returned %d devices, want %d", len(devices), len(testDevices))
	}

	// Verify all devices are present
	deviceMap := make(map[string]bool)
	for _, d := range devices {
		deviceMap[d.ID] = true
	}
	for _, d := range testDevices {
		if !deviceMap[d.id] {
			t.Errorf("List() missing device with ID %s", d.id)
		}
	}
}

func TestConcurrentOperations(t *testing.T) {
	store := NewInMemoryStore()
	numGoroutines := 10
	var wg sync.WaitGroup

	// Test concurrent creation
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			id := fmt.Sprintf("device-%d", index)
			_, err := store.Create(id, domain.ECC, "Test Device")
			if err != nil {
				t.Errorf("Failed to create device in goroutine %d: %v", index, err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all devices were created
	devices, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(devices) != numGoroutines {
		t.Errorf("Expected %d devices, got %d", numGoroutines, len(devices))
	}

	// Test concurrent reads
	wg.Add(numGoroutines * 2) // Each device will be read twice
	for i := 0; i < numGoroutines; i++ {
		id := fmt.Sprintf("device-%d", i)
		go func(deviceID string) {
			defer wg.Done()
			_, err := store.Get(deviceID)
			if err != nil {
				t.Errorf("Failed to get device %s: %v", deviceID, err)
			}
		}(id)
		go func(deviceID string) {
			defer wg.Done()
			_, err := store.Get(deviceID)
			if err != nil {
				t.Errorf("Failed to get device %s: %v", deviceID, err)
			}
		}(id)
	}
	wg.Wait()
}
