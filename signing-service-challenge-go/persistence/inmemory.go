package persistence

import (
	"sync"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
)

// InMemoryStore implements the DeviceStore interface using an in-memory map
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

// Create stores a new signature device
func (store *InMemoryStore) Create(device *domain.SignatureDevice) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.devices[device.ID]; exists {
		return domain.ErrDeviceAlreadyExists
	}

	store.devices[device.ID] = device
	return nil
}

// Get retrieves a signature device by ID
func (store *InMemoryStore) Get(deviceID string) (*domain.SignatureDevice, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	device, exists := store.devices[deviceID]
	if !exists {
		return nil, domain.ErrDeviceNotFound
	}

	return device, nil
}

// List returns all stored signature devices
func (store *InMemoryStore) List() ([]*domain.SignatureDevice, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	devices := make([]*domain.SignatureDevice, 0, len(store.devices))
	for _, device := range store.devices {
		devices = append(devices, device)
	}

	return devices, nil
}
