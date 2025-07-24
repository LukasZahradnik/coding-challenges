package api

import (
	"encoding/json"
	"net/http"

	"github.com/fiskaly/coding-challenges/signing-service-challenge/domain"
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
	store         persistence.Store[domain.SignatureDevice]
}

// NewServer is a factory to instantiate a new Server.
func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		store:         persistence.NewInMemoryStore[domain.SignatureDevice](),
	}
}

// Run registers all HandlerFuncs for the existing HTTP routes and starts the Server.
func (s *Server) Run() error {
	mux := http.NewServeMux()

	mux.Handle("/api/v0/health", http.HandlerFunc(s.Health))
	mux.Handle("/api/v0/devices", handlersWithMethods(map[string]http.HandlerFunc{
		http.MethodPost: http.HandlerFunc(s.CreateSignatureDevice),
		http.MethodGet:  http.HandlerFunc(s.ListSignatureDevice),
	}))

	mux.Handle("/api/v0/devices/{id}", handlersWithMethods(map[string]http.HandlerFunc{
		http.MethodGet: http.HandlerFunc(s.GetSignatureDevice)},
	))

	mux.Handle("/api/v0/devices/{id}/sign", handlersWithMethods(map[string]http.HandlerFunc{
		http.MethodPost: http.HandlerFunc(s.SignTransaction)},
	))

	return http.ListenAndServe(s.listenAddress, mux)
}

func handlersWithMethods(handlerFuncs map[string]http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, ok := handlerFuncs[r.Method]
		if !ok {
			WriteErrorResponse(w, http.StatusMethodNotAllowed, []string{
				http.StatusText(http.StatusMethodNotAllowed),
			})

			return
		}

		handler(w, r)
	})
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes)
}
