package kaboo

import (
	"fmt"
	"github.com/auth0-community/go-auth0"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	jose "gopkg.in/square/go-jose.v2"
	"net/http"
	"os"
)

// Server main kaboo server
type Server struct {
	RESTPort      int
	WebSocketPort int
	Auth0Domain   string
	Auth0Audience string
}

// NewServer initializes a new kaboo server
func NewServer(restPort int, wsPort int, auth0Domain string, auth0Audience string) Server {
	return Server{restPort, wsPort, auth0Domain, auth0Audience}
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: fmt.Sprintf("https://%s/.well-known/jwks.json", s.Auth0Domain)}, nil)
		audience := s.Auth0Audience
		configuration := auth0.NewConfiguration(client, []string{audience}, fmt.Sprintf("https://%s/", s.Auth0Domain), jose.RS256)
		validator := auth0.NewValidator(configuration, nil)

		token, err := validator.ValidateRequest(r)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Token is not valid:", token)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// Start starts the server
func (s *Server) Start() {
	var NotImplemented = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Not Implemented"))
	})

	r := mux.NewRouter()
	r.Handle("/api/game/new", s.authMiddleware(NotImplemented))
	r.Handle("/api/game/join", s.authMiddleware(NotImplemented))
	r.Handle("/api/game/leave", s.authMiddleware(NotImplemented))
	http.ListenAndServe(fmt.Sprintf(":%d", s.RESTPort), handlers.LoggingHandler(os.Stdout, r))
}
