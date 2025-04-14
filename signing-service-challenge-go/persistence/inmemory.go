package persistence

import (
	"errors"
	"sync"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
)

var (
	// ErrDeviceNotFound is returned when a device with the given ID doesn't exist
	ErrDeviceNotFound = errors.New("device not found")
	// ErrDeviceAlreadyExists is returned when trying to create a device with an ID that already exists
	ErrDeviceAlreadyExists = errors.New("device already exists")
)

// InMemoryStore is an in-memory implementation of the device storage
type InMemoryStore struct {
	devices map[string]*domain.SignatureDevice
	mu      sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		devices: make(map[string]*domain.SignatureDevice),
	}
}

// CreateDevice stores a new signature device
func (s *InMemoryStore) CreateDevice(device *domain.SignatureDevice) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[device.ID]; exists {
		return ErrDeviceAlreadyExists
	}

	s.devices[device.ID] = device
	return nil
}

// GetDevice retrieves a signature device by its ID
func (s *InMemoryStore) GetDevice(id string) (*domain.SignatureDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, ErrDeviceNotFound
	}

	return device, nil
}

// ListDevices returns all stored signature devices
func (s *InMemoryStore) ListDevices() []*domain.SignatureDevice {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]*domain.SignatureDevice, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}

	return devices
}
