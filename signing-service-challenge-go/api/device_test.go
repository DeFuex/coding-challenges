package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
)

func TestCreateDevice(t *testing.T) {
	store := persistence.NewInMemoryStore()
	handler := NewDeviceHandler(store)

	tests := []struct {
		name       string
		request    DeviceRequest
		wantStatus int
		wantError  bool
	}{
		{
			name: "Valid RSA device",
			request: DeviceRequest{
				ID:        "test-rsa",
				Algorithm: domain.RSA,
				Label:     "Test RSA Device",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "Valid ECC device",
			request: DeviceRequest{
				ID:        "test-ecc",
				Algorithm: domain.ECC,
				Label:     "Test ECC Device",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "Invalid algorithm",
			request: DeviceRequest{
				ID:        "test-invalid",
				Algorithm: "INVALID",
				Label:     "Test Invalid Device",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "Duplicate device ID",
			request: DeviceRequest{
				ID:        "test-rsa", // Same as first test
				Algorithm: domain.RSA,
				Label:     "Duplicate Device",
			},
			wantStatus: http.StatusConflict,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v0/signature-devices", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.CreateDevice(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateDevice() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantError {
				var errResp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Errorf("Failed to decode error response: %v", err)
				}
				if len(errResp.Errors) == 0 {
					t.Error("Expected error message in response")
				}
			} else {
				var resp Response
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Errorf("Failed to decode success response: %v", err)
				}

				var device DeviceResponse
				deviceData, err := json.Marshal(resp.Data)
				if err != nil {
					t.Errorf("Failed to marshal device data: %v", err)
				}
				if err := json.Unmarshal(deviceData, &device); err != nil {
					t.Errorf("Failed to unmarshal device data: %v", err)
				}

				if device.ID != tt.request.ID {
					t.Errorf("Response ID = %v, want %v", device.ID, tt.request.ID)
				}
				if device.Algorithm != tt.request.Algorithm {
					t.Errorf("Response Algorithm = %v, want %v", device.Algorithm, tt.request.Algorithm)
				}
				if device.Label != tt.request.Label {
					t.Errorf("Response Label = %v, want %v", device.Label, tt.request.Label)
				}
				if device.SignatureCounter != 0 {
					t.Errorf("Initial SignatureCounter = %v, want 0", device.SignatureCounter)
				}
				if device.PublicKey == "" {
					t.Error("Public key is empty")
				}
			}
		})
	}
}

func TestGetDevice(t *testing.T) {
	store := persistence.NewInMemoryStore()
	handler := NewDeviceHandler(store)

	// Create a test device first
	createReq := DeviceRequest{
		ID:        "test-device",
		Algorithm: domain.RSA,
		Label:     "Test Device",
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/signature-devices", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateDevice(rec, req)

	tests := []struct {
		name       string
		deviceID   string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "Existing device",
			deviceID:   "test-device",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "Non-existent device",
			deviceID:   "non-existent",
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
		{
			name:       "Empty device ID",
			deviceID:   "",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v0/signature-devices/get?id="+tt.deviceID, nil)
			rec := httptest.NewRecorder()

			handler.GetDevice(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetDevice() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantError {
				var errResp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Errorf("Failed to decode error response: %v", err)
				}
				if len(errResp.Errors) == 0 {
					t.Error("Expected error message in response")
				}
			} else {
				var resp Response
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Errorf("Failed to decode success response: %v", err)
				}

				var device DeviceResponse
				deviceData, err := json.Marshal(resp.Data)
				if err != nil {
					t.Errorf("Failed to marshal device data: %v", err)
				}
				if err := json.Unmarshal(deviceData, &device); err != nil {
					t.Errorf("Failed to unmarshal device data: %v", err)
				}

				if device.ID != tt.deviceID {
					t.Errorf("Response ID = %v, want %v", device.ID, tt.deviceID)
				}
			}
		})
	}
}

func TestListDevices(t *testing.T) {
	store := persistence.NewInMemoryStore()
	handler := NewDeviceHandler(store)

	// Create test devices
	devices := []DeviceRequest{
		{
			ID:        "device-1",
			Algorithm: domain.RSA,
			Label:     "Device 1",
		},
		{
			ID:        "device-2",
			Algorithm: domain.ECC,
			Label:     "Device 2",
		},
	}

	for _, device := range devices {
		body, _ := json.Marshal(device)
		req := httptest.NewRequest(http.MethodPost, "/api/v0/signature-devices", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		handler.CreateDevice(rec, req)
	}

	// Test listing devices
	req := httptest.NewRequest(http.MethodGet, "/api/v0/signature-devices/list", nil)
	rec := httptest.NewRecorder()

	handler.ListDevices(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ListDevices() status = %v, want %v", rec.Code, http.StatusOK)
	}

	var resp Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	var deviceList []DeviceResponse
	deviceData, err := json.Marshal(resp.Data)
	if err != nil {
		t.Errorf("Failed to marshal device data: %v", err)
	}
	if err := json.Unmarshal(deviceData, &deviceList); err != nil {
		t.Errorf("Failed to unmarshal device data: %v", err)
	}

	if len(deviceList) != len(devices) {
		t.Errorf("ListDevices() returned %v devices, want %v", len(deviceList), len(devices))
	}

	// Verify each device in the response
	deviceMap := make(map[string]DeviceResponse)
	for _, device := range deviceList {
		deviceMap[device.ID] = device
	}

	for _, want := range devices {
		got, exists := deviceMap[want.ID]
		if !exists {
			t.Errorf("Device %v not found in response", want.ID)
			continue
		}
		if got.Algorithm != want.Algorithm {
			t.Errorf("Device %v algorithm = %v, want %v", want.ID, got.Algorithm, want.Algorithm)
		}
		if got.Label != want.Label {
			t.Errorf("Device %v label = %v, want %v", want.ID, got.Label, want.Label)
		}
	}
}

func TestSignTransaction(t *testing.T) {
	store := persistence.NewInMemoryStore()
	handler := NewDeviceHandler(store)

	// Create a test device
	createReq := DeviceRequest{
		ID:        "test-device",
		Algorithm: domain.RSA,
		Label:     "Test Device",
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/signature-devices", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.CreateDevice(rec, req)

	tests := []struct {
		name       string
		deviceID   string
		data       string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "Valid transaction",
			deviceID:   "test-device",
			data:       "test data",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "Non-existent device",
			deviceID:   "non-existent",
			data:       "test data",
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
		{
			name:       "Empty device ID",
			deviceID:   "",
			data:       "test data",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "Empty data",
			deviceID:   "test-device",
			data:       "",
			wantStatus: http.StatusOK, // Empty data is allowed
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signReq := SignTransactionRequest{
				Data: tt.data,
			}
			body, _ := json.Marshal(signReq)
			req := httptest.NewRequest(http.MethodPost, "/api/v0/signature-devices/sign?id="+tt.deviceID, bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.SignTransaction(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("SignTransaction() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantError {
				var errResp ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Errorf("Failed to decode error response: %v", err)
				}
				if len(errResp.Errors) == 0 {
					t.Error("Expected error message in response")
				}
			} else {
				var resp Response
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Errorf("Failed to decode success response: %v", err)
				}

				var signature map[string]string
				sigData, err := json.Marshal(resp.Data)
				if err != nil {
					t.Errorf("Failed to marshal signature data: %v", err)
				}
				if err := json.Unmarshal(sigData, &signature); err != nil {
					t.Errorf("Failed to unmarshal signature data: %v", err)
				}

				if signature["signature"] == "" {
					t.Error("Signature is empty")
				}
				if signature["signed_data"] == "" {
					t.Error("Signed data is empty")
				}
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	store := persistence.NewInMemoryStore()
	handler := NewDeviceHandler(store)

	tests := []struct {
		name     string
		method   string
		endpoint string
		handler  func(http.ResponseWriter, *http.Request)
	}{
		{
			name:     "Create device with GET",
			method:   http.MethodGet,
			endpoint: "/api/v0/signature-devices",
			handler:  handler.CreateDevice,
		},
		{
			name:     "Get device with POST",
			method:   http.MethodPost,
			endpoint: "/api/v0/signature-devices/",
			handler:  handler.GetDevice,
		},
		{
			name:     "List devices with POST",
			method:   http.MethodPost,
			endpoint: "/api/v0/signature-devices/list",
			handler:  handler.ListDevices,
		},
		{
			name:     "Sign transaction with GET",
			method:   http.MethodGet,
			endpoint: "/api/v0/signature-devices/sign",
			handler:  handler.SignTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			rec := httptest.NewRecorder()

			tt.handler(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%v status = %v, want %v", tt.name, rec.Code, http.StatusMethodNotAllowed)
			}

			var errResp ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
				t.Errorf("Failed to decode error response: %v", err)
			}
			if len(errResp.Errors) == 0 {
				t.Error("Expected error message in response")
			}
		})
	}
}
