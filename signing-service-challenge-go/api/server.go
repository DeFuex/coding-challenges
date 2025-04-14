package api

import (
	"encoding/json"
	"net/http"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/persistence"
)

// Response is the generic API response container.
type Response struct {
	Data interface{} `json:"data"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}

// Server manages HTTP requests and dispatches them to the appropriate services.
type Server struct {
	listenAddress string
	deviceHandler *DeviceHandler
}

// NewServer is a factory to instantiate a new Server.
func NewServer(listenAddress string) *Server {
	store := persistence.NewInMemoryStore()
	deviceHandler := NewDeviceHandler(store)

	return &Server{
		listenAddress: listenAddress,
		deviceHandler: deviceHandler,
	}
}

// Run registers all HandlerFuncs for the existing HTTP routes and starts the Server.
func (server *Server) Run() error {
	mux := http.NewServeMux()

	// Health endpoint
	mux.Handle("/api/v0/health", http.HandlerFunc(server.Health))

	// Device endpoints
	mux.Handle("/api/v0/signature-devices", http.HandlerFunc(server.deviceHandler.CreateDevice))
	mux.Handle("/api/v0/signature-devices/get", http.HandlerFunc(server.deviceHandler.GetDevice))
	mux.Handle("/api/v0/signature-devices/list", http.HandlerFunc(server.deviceHandler.ListDevices))
	mux.Handle("/api/v0/signature-devices/sign", http.HandlerFunc(server.deviceHandler.SignTransaction))
	mux.Handle("/api/v0/signature-devices/delete", http.HandlerFunc(server.deviceHandler.DeleteDevice))
	mux.Handle("/api/v0/signature-devices/update", http.HandlerFunc(server.deviceHandler.UpdateDevice))

	return http.ListenAndServe(server.listenAddress, mux)
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteAPIResponse writes a successful API response.
func WriteAPIResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
		return
	}

	w.Write(bytes)
}

// WriteErrorResponse writes an error API response.
func WriteErrorResponse(w http.ResponseWriter, statusCode int, errors []string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
		return
	}

	w.Write(bytes)
}
