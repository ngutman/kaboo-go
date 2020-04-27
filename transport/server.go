package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/ngutman/kaboo-server-go/transport/websocket"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ngutman/kaboo-server-go/api/types"
	"github.com/ngutman/kaboo-server-go/backend"
	"github.com/ngutman/kaboo-server-go/models"
)

// Server main kaboo server
type Server struct {
	authMiddleware JWTAuthMiddleware
	api            API
	hub            *websocket.Hub
	restPort       int
}

// API wires incoming requests to their respective backend engines
type API struct {
	gameBackend types.GameBackend
}

// NewServer initializes a new kaboo server
func NewServer(restPort int, auth0Domain string, auth0Audience string) Server {
	var db models.Db
	db.Open("mongodb://localhost:27017/", "kaboo")
	hub := websocket.NewHub()
	go hub.Run()
	return Server{
		JWTAuthMiddleware{
			&db,
			auth0Domain,
			auth0Audience,
		},
		API{
			gameBackend: backend.NewGameController(&db, hub),
		},
		hub,
		restPort,
	}
}

// Start starts the server
func (s *Server) Start() {
	r := mux.NewRouter()
	if os.Getenv("DEBUG") != "" {
		r.Use(handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
			handlers.AllowCredentials(),
		))
	}
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/game/new", s.authMiddleware.Handle(s.api.handleNewGame))
	apiRouter.HandleFunc("/game/join", s.authMiddleware.Handle(s.api.handleJoinGame))
	apiRouter.HandleFunc("/game/leave", s.authMiddleware.Handle(s.api.handleLeaveGame))

	apiRouter.HandleFunc("/state", s.authMiddleware.Handle(notImplemented))
	apiRouter.HandleFunc("/ws", s.authMiddleware.Handle(s.hub.HandleWSUpgradeRequest))

	log.Infof("Starting API server (:%v)\n", s.restPort)
	http.ListenAndServe(fmt.Sprintf(":%d", s.restPort), handlers.CombinedLoggingHandler(log.StandardLogger().Out, r))
}

func (a *API) handleNewGame(w http.ResponseWriter, r *http.Request, user *models.User) {
	var req createGameReq
	if tryToDecodeOrFail(w, r, &req) != nil {
		return
	}
	gameID, err := a.gameBackend.NewGame(user, req.Name, req.MaxPlayersCount, req.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tryToWriteJSONResponse(w, r, &createGameRes{GameID: gameID})
}

func (a *API) handleJoinGame(w http.ResponseWriter, r *http.Request, user *models.User) {
	var req joinGameReq
	if tryToDecodeOrFail(w, r, &req) != nil {
		return
	}
	success, err := a.gameBackend.JoinGameByGameID(user, req.GameID, req.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tryToWriteJSONResponse(w, r, &joinGameRes{Success: success})
}

func (a *API) handleLeaveGame(w http.ResponseWriter, r *http.Request, user *models.User) {
	notImplemented(w, r, user)
}

func tryToDecodeOrFail(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if err := decodeJSONBody(w, r, dst); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return err
	}
	return nil
}

func tryToWriteJSONResponse(w http.ResponseWriter, r *http.Request, res interface{}) error {
	jsonRes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Write([]byte(jsonRes))
	return nil
}

func notImplemented(w http.ResponseWriter, r *http.Request, user *models.User) {
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}
