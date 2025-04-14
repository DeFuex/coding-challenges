package api

import (
	"encoding/json"
	"net/http"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
)

// DeviceRequest represents the request body for creating a device
type DeviceRequest struct {
	ID        string                    `json:"id"`
	Algorithm domain.SignatureAlgorithm `json:"algorithm"`
	Label     string                    `json:"label"`
}

// DeviceResponse represents the response body for a device
type DeviceResponse struct {
	ID               string                    `json:"id"`
	Label            string                    `json:"label"`
	Algorithm        domain.SignatureAlgorithm `json:"algorithm"`
	SignatureCounter int                       `json:"signature_counter"`
	LastSignature    string                    `json:"last_signature"`
	PublicKey        string                    `json:"public_key"`
}

// SignTransactionRequest represents the request body for signing a transaction
type SignTransactionRequest struct {
	Data string `json:"data"`
}

// UpdateRequest represents the request body for updating a device
type UpdateRequest struct {
	Label string `json:"label"`
}

// DeviceHandler handles device-related requests
type DeviceHandler struct {
	store *persistence.InMemoryStore
}

// NewDeviceHandler creates a new DeviceHandler
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

	var req DeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"invalid request body"})
		return
	}

	device, err := h.store.Create(req.ID, req.Algorithm, req.Label)
	if err != nil {
		switch err {
		case domain.ErrDeviceAlreadyExists:
			WriteErrorResponse(w, http.StatusConflict, []string{"device already exists"})
		case domain.ErrInvalidAlgorithm:
			WriteErrorResponse(w, http.StatusBadRequest, []string{"invalid algorithm"})
		default:
			WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to create device"})
		}
		return
	}

	WriteAPIResponse(w, http.StatusCreated, DeviceResponse{
		ID:               device.ID,
		Label:            device.Label,
		Algorithm:        device.Algorithm,
		SignatureCounter: device.SignatureCounter,
		LastSignature:    device.LastSignature,
		PublicKey:        string(device.PublicKey),
	})
}

// GetDevice handles the retrieval of a signature device
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

	device, err := h.store.Get(deviceID)
	if err != nil {
		if err == domain.ErrDeviceNotFound {
			WriteErrorResponse(w, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to get device"})
		return
	}

	WriteAPIResponse(w, http.StatusOK, DeviceResponse{
		ID:               device.ID,
		Label:            device.Label,
		Algorithm:        device.Algorithm,
		SignatureCounter: device.SignatureCounter,
		LastSignature:    device.LastSignature,
		PublicKey:        string(device.PublicKey),
	})
}

// ListDevices handles the retrieval of all signature devices
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	devices, err := h.store.List()
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to list devices"})
		return
	}

	response := make([]DeviceResponse, len(devices))
	for i, device := range devices {
		response[i] = DeviceResponse{
			ID:               device.ID,
			Label:            device.Label,
			Algorithm:        device.Algorithm,
			SignatureCounter: device.SignatureCounter,
			LastSignature:    device.LastSignature,
			PublicKey:        string(device.PublicKey),
		}
	}

	WriteAPIResponse(w, http.StatusOK, response)
}

// SignTransaction handles the signing of transaction data
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

	device, err := h.store.Get(deviceID)
	if err != nil {
		if err == domain.ErrDeviceNotFound {
			WriteErrorResponse(w, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to get device"})
		return
	}

	signature, err := device.SignTransaction(req.Data)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to sign transaction"})
		return
	}

	WriteAPIResponse(w, http.StatusOK, signature)
}

// DeleteDevice handles the deletion of a signature device
func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	deviceID := r.URL.Query().Get("id")
	if deviceID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"device ID is required"})
		return
	}

	err := h.store.Delete(deviceID)
	if err != nil {
		if err == domain.ErrDeviceNotFound {
			WriteErrorResponse(w, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to delete device"})
		return
	}

	WriteAPIResponse(w, http.StatusOK, map[string]string{"message": "device deleted successfully"})
}

// UpdateDevice handles updating a signature device's label
func (h *DeviceHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	deviceID := r.URL.Query().Get("id")
	if deviceID == "" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"device ID is required"})
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"invalid request body"})
		return
	}

	if req.Label == "" {
		WriteErrorResponse(w, http.StatusBadRequest, []string{"label is required"})
		return
	}

	device, err := h.store.Update(deviceID, req.Label)
	if err != nil {
		if err == domain.ErrDeviceNotFound {
			WriteErrorResponse(w, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteErrorResponse(w, http.StatusInternalServerError, []string{"failed to update device"})
		return
	}

	WriteAPIResponse(w, http.StatusOK, DeviceResponse{
		ID:               device.ID,
		Label:            device.Label,
		Algorithm:        device.Algorithm,
		SignatureCounter: device.SignatureCounter,
		LastSignature:    device.LastSignature,
		PublicKey:        string(device.PublicKey),
	})
}
