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
	mux.Handle("/api/v1/health", http.HandlerFunc(server.Health))

	// Device endpoints
	mux.Handle("/api/v1/signature-devices", http.HandlerFunc(server.deviceHandler.CreateDevice))
	mux.Handle("/api/v1/signature-devices/", http.HandlerFunc(server.deviceHandler.GetDevice))
	mux.Handle("/api/v1/signature-devices/list", http.HandlerFunc(server.deviceHandler.ListDevices))
	mux.Handle("/api/v1/signature-devices/sign", http.HandlerFunc(server.deviceHandler.SignTransaction))

	return http.ListenAndServe(server.listenAddress, mux)
}

// Health handles the health check endpoint
func (server *Server) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{"method not allowed"})
		return
	}

	WriteAPIResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(responseWriter http.ResponseWriter) {
	responseWriter.WriteHeader(http.StatusInternalServerError)
	responseWriter.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(responseWriter http.ResponseWriter, statusCode int, errors []string) {
	responseWriter.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	responseBytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(responseWriter)
	}

	responseWriter.Write(responseBytes)
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(responseWriter http.ResponseWriter, statusCode int, data interface{}) {
	responseWriter.WriteHeader(statusCode)

	response := Response{
		Data: data,
	}

	responseBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(responseWriter)
	}

	responseWriter.Write(responseBytes)
}
