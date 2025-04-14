package api

import (
	"encoding/json"
	"net/http"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/crypto"
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
func (handler *DeviceHandler) CreateDevice(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		WriteErrorResponse(responseWriter, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	var createRequest CreateDeviceRequest
	if err := json.NewDecoder(request.Body).Decode(&createRequest); err != nil {
		WriteErrorResponse(responseWriter, http.StatusBadRequest, []string{"invalid request body"})
		return
	}

	var signer crypto.Signer
	var publicKey, privateKey []byte

	switch createRequest.Algorithm {
	case domain.ECC:
		// Generate ECC key pair
		eccGenerator := &crypto.ECCGenerator{}
		eccKeyPair, err := eccGenerator.Generate()
		if err != nil {
			WriteErrorResponse(responseWriter, http.StatusInternalServerError, []string{"failed to generate ECC key pair"})
			return
		}

		// Marshal keys
		eccMarshaler := crypto.NewECCMarshaler()
		publicKey, privateKey, err = eccMarshaler.Encode(*eccKeyPair)
		if err != nil {
			WriteErrorResponse(responseWriter, http.StatusInternalServerError, []string{"failed to marshal ECC keys"})
			return
		}

		// Create ECC signer
		signer = crypto.NewECCSigner(eccKeyPair.Private)

	case domain.RSA:
		// Generate RSA key pair
		rsaGenerator := &crypto.RSAGenerator{}
		rsaKeyPair, err := rsaGenerator.Generate()
		if err != nil {
			WriteErrorResponse(responseWriter, http.StatusInternalServerError, []string{"failed to generate RSA key pair"})
			return
		}

		// Marshal keys
		rsaMarshaler := crypto.NewRSAMarshaler()
		publicKey, privateKey, err = rsaMarshaler.Marshal(*rsaKeyPair)
		if err != nil {
			WriteErrorResponse(responseWriter, http.StatusInternalServerError, []string{"failed to marshal RSA keys"})
			return
		}

		// Create RSA signer
		signer = crypto.NewRSASigner(rsaKeyPair.Private)

	default:
		WriteErrorResponse(responseWriter, http.StatusBadRequest, []string{"unsupported algorithm"})
		return
	}

	// Create the device
	device := domain.NewSignatureDevice(
		createRequest.ID,
		createRequest.Algorithm,
		createRequest.Label,
		signer,
		privateKey,
		publicKey,
	)

	if err := handler.store.CreateDevice(device); err != nil {
		if err == persistence.ErrDeviceAlreadyExists {
			WriteErrorResponse(responseWriter, http.StatusConflict, []string{"device already exists"})
			return
		}
		WriteInternalError(responseWriter)
		return
	}

	WriteAPIResponse(responseWriter, http.StatusCreated, device)
}

// GetDevice handles retrieving a signature device by ID
func (handler *DeviceHandler) GetDevice(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		WriteErrorResponse(responseWriter, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	deviceID := request.URL.Query().Get("id")
	if deviceID == "" {
		WriteErrorResponse(responseWriter, http.StatusBadRequest, []string{"device ID is required"})
		return
	}

	device, err := handler.store.GetDevice(deviceID)
	if err != nil {
		if err == persistence.ErrDeviceNotFound {
			WriteErrorResponse(responseWriter, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteInternalError(responseWriter)
		return
	}

	WriteAPIResponse(responseWriter, http.StatusOK, device)
}

// ListDevices handles retrieving all signature devices
func (handler *DeviceHandler) ListDevices(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		WriteErrorResponse(responseWriter, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	devices := handler.store.ListDevices()
	WriteAPIResponse(responseWriter, http.StatusOK, devices)
}

// SignTransaction handles signing a transaction with a signature device
func (handler *DeviceHandler) SignTransaction(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		WriteErrorResponse(responseWriter, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	deviceID := request.URL.Query().Get("id")
	if deviceID == "" {
		WriteErrorResponse(responseWriter, http.StatusBadRequest, []string{"device ID is required"})
		return
	}

	var signRequest SignTransactionRequest
	if err := json.NewDecoder(request.Body).Decode(&signRequest); err != nil {
		WriteErrorResponse(responseWriter, http.StatusBadRequest, []string{"invalid request body"})
		return
	}

	device, err := handler.store.GetDevice(deviceID)
	if err != nil {
		if err == persistence.ErrDeviceNotFound {
			WriteErrorResponse(responseWriter, http.StatusNotFound, []string{"device not found"})
			return
		}
		WriteInternalError(responseWriter)
		return
	}

	signature, err := device.SignTransaction(signRequest.Data)
	if err != nil {
		WriteErrorResponse(responseWriter, http.StatusInternalServerError, []string{"failed to sign transaction"})
		return
	}

	WriteAPIResponse(responseWriter, http.StatusOK, signature)
}
