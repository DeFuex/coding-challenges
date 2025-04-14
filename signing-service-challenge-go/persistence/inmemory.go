package persistence

import (
	"sync"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/crypto"
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

// Create creates a new signature device
func (s *InMemoryStore) Create(id string, algorithm domain.SignatureAlgorithm, label string) (*domain.SignatureDevice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[id]; exists {
		return nil, domain.ErrDeviceAlreadyExists
	}

	var signer crypto.Signer
	var privateKey, publicKey []byte
	var err error

	switch algorithm {
	case domain.ECC:
		signer, privateKey, publicKey, err = crypto.NewECCSigner()
	case domain.RSA:
		signer, privateKey, publicKey, err = crypto.NewRSASigner()
	default:
		return nil, domain.ErrInvalidAlgorithm
	}

	if err != nil {
		return nil, err
	}

	device := domain.NewSignatureDevice(id, algorithm, label, signer, privateKey, publicKey)
	s.devices[id] = device

	return device, nil
}

// Get retrieves a signature device by ID
func (s *InMemoryStore) Get(id string) (*domain.SignatureDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, domain.ErrDeviceNotFound
	}

	return device, nil
}

// List retrieves all signature devices
func (s *InMemoryStore) List() ([]*domain.SignatureDevice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]*domain.SignatureDevice, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

// Delete removes a signature device by ID
func (s *InMemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.devices[id]; !exists {
		return domain.ErrDeviceNotFound
	}

	delete(s.devices, id)
	return nil
}

// Update updates a signature device's label
func (s *InMemoryStore) Update(id string, label string) (*domain.SignatureDevice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, exists := s.devices[id]
	if !exists {
		return nil, domain.ErrDeviceNotFound
	}

	device.Label = label
	return device, nil
}
