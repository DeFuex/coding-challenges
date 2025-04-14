package api

import (
	"encoding/json"
	"net/http"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
)

// CreateDeviceRequest represents the request body for creating a new signature device
type CreateDeviceRequest struct {
	ID        string                    `json:"id"`
	Algorithm domain.SignatureAlgorithm `json:"algorithm"`
	Label     string                    `json:"label,omitempty"`
}

// SignTransactionRequest represents the request body for signing a transaction
type SignTransactionRequest struct {
	Data string `json:"data"`
}

// DeviceHandler handles HTTP requests for signature devices
type DeviceHandler struct {
	store *persistence.InMemoryStore
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(store *persistence.InMemoryStore) *DeviceHandler {
	return &DeviceHandler{
		store: store,
	}
}

// CreateDevice handles the creation of a new signature device
func (h *DeviceHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	var req CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"invalid request body"})
		return
	}

	// TODO: Generate key pair and create signer based on algorithm
	// For now, we'll just create a device with empty keys
	device := domain.NewSignatureDevice(
		req.ID,
		req.Algorithm,
		req.Label,
		nil, // TODO: Create proper signer
		nil, // TODO: Add private key
		nil, // TODO: Add public key
	)

	if err := h.store.CreateDevice(device); err != nil {
		if err == persistence.ErrDeviceAlreadyExists {
			WriteErrorResponse(w, http.StatusConflict, []string{"device already exists"})
			return
		}
		WriteInternalError(w)
		return
	}

	WriteAPIResponse(w, http.StatusCreated, device)
}

// GetDevice handles retrieving a signature device by ID
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	deviceID := r.URL.Query().Get("id")
	if deviceID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"device ID is required"})
		return
	}

	device, err := h.store.GetDevice(deviceID)
	if err != nil {
		if err == persistence.ErrDeviceNotFound {
			WriteErrorResponse(w, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteInternalError(w)
		return
	}

	WriteAPIResponse(w, http.StatusOK, device)
}

// ListDevices handles retrieving all signature devices
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	devices := h.store.ListDevices()
	WriteAPIResponse(w, http.StatusOK, devices)
}

// SignTransaction handles signing a transaction with a signature device
func (h *DeviceHandler) SignTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	deviceID := r.URL.Query().Get("id")
	if deviceID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"device ID is required"})
		return
	}

	var req SignTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"invalid request body"})
		return
	}

	device, err := h.store.GetDevice(deviceID)
	if err != nil {
		if err == persistence.ErrDeviceNotFound {
			WriteErrorResponse(w, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteInternalError(w)
		return
	}

	signature, err := device.SignTransaction(req.Data)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to sign transaction"})
		return
	}

	WriteAPIResponse(w, http.StatusOK, signature)
}
