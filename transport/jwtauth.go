package transport

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/auth0-community/go-auth0"
	"github.com/ngutman/kaboo-server-go/models"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// JWTAuthMiddleware handles Auth0 JWT validation
type JWTAuthMiddleware struct {
	db            *models.Db
	auth0Domain   string
	auth0Audience string
}

// Handle implements the JWT validation over incoming request
func (j *JWTAuthMiddleware) Handle(next func(w http.ResponseWriter, r *http.Request, user *models.User)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// For faster development allowed to skip token authentication in DEBUG mode
		if os.Getenv("DEBUG") != "" {
			debugToken := regexp.MustCompile(`DebugToken (.*)`).FindStringSubmatch(r.Header.Get("Authorization"))
			if len(debugToken) > 1 {
				log.Debugf("Debug token, setting user to %v\n", debugToken[1])
				user, _ := j.db.UserDAO.FetchUserByExternalID(debugToken[1])
				next(w, r, user)
				return
			}
		}
		client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: fmt.Sprintf("https://%s/.well-known/jwks.json", j.auth0Domain)}, nil)
		audience := j.auth0Audience
		configuration := auth0.NewConfiguration(client, []string{audience}, fmt.Sprintf("https://%s/", j.auth0Domain), jose.RS256)
		validator := auth0.NewValidator(configuration, nil)

		token, err := validator.ValidateRequest(r)
		if err != nil {
			log.Errorf("Token %v is invalid, %v\n", token, err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			claims := jwt.Claims{}
			validator.Claims(r, token, &claims)
			// Attach user to request context
			user, _ := j.db.UserDAO.FetchUserByExternalID(claims.Subject)
			next(w, r, user)
		}
	}
}
