package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	server := NewServer(":8080")

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "GET request",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "POST request",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  true,
		},
		{
			name:       "PUT request",
			method:     http.MethodPut,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  true,
		},
		{
			name:       "DELETE request",
			method:     http.MethodDelete,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v0/health", nil)
			rec := httptest.NewRecorder()

			server.Health(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Health() status = %v, want %v", rec.Code, tt.wantStatus)
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

				health, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Error("Response data is not a map")
					return
				}

				status, ok := health["status"].(string)
				if !ok {
					t.Error("Status is not a string")
					return
				}
				if status != "pass" {
					t.Errorf("Status = %v, want pass", status)
				}

				version, ok := health["version"].(string)
				if !ok {
					t.Error("Version is not a string")
					return
				}
				if version != "v0" {
					t.Errorf("Version = %v, want v0", version)
				}
			}

			contentType := rec.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}
		})
	}
}
