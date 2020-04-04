package transport

import (
	"fmt"
	"github.com/auth0-community/go-auth0"
	"gopkg.in/square/go-jose.v2"
	"net/http"
)

// JWTAuthMiddleware handles Auth0 JWT validation
type JWTAuthMiddleware struct {
	auth0Domain   string
	auth0Audience string
}

// Handle implements the JWT validation over incoming request
func (j *JWTAuthMiddleware) Handle(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: fmt.Sprintf("https://%s/.well-known/jwks.json", j.auth0Domain)}, nil)
		audience := j.auth0Audience
		configuration := auth0.NewConfiguration(client, []string{audience}, fmt.Sprintf("https://%s/", j.auth0Domain), jose.RS256)
		validator := auth0.NewValidator(configuration, nil)

		token, err := validator.ValidateRequest(r)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Token is not valid:", token)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			next(w, r)
		}
	}
}