package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ngutman/kaboo-server-go/service"
	"log"
	"net/http"
	"os"
)

// Server main kaboo server
type Server struct {
	authMiddleware JWTAuthMiddleware
	api            API
	restPort       int
	webSocketPort  int
}

type API struct {
	gameController service.GameController
}

// NewServer initializes a new kaboo server
func NewServer(restPort int, wsPort int, auth0Domain string, auth0Audience string) Server {
	return Server{
		JWTAuthMiddleware{auth0Domain, auth0Audience},
		API{
			service.GameController{},
		},
		restPort,
		wsPort,
	}
}

// Start starts the server
func (s *Server) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/api/game/new", s.authMiddleware.Handle(s.api.handleNewGame))
	r.HandleFunc("/api/game/join", s.authMiddleware.Handle(s.api.handleJoinGame))
	r.HandleFunc("/api/game/leave", s.authMiddleware.Handle(s.api.handleLeaveGame))

	r.HandleFunc("/api/state", s.authMiddleware.Handle(notImplemented))
	r.HandleFunc("/api/ws", s.authMiddleware.Handle(notImplemented))

	http.ListenAndServe(fmt.Sprintf(":%d", s.restPort), handlers.CombinedLoggingHandler(os.Stdout, r))
}

func (a *API) handleNewGame(w http.ResponseWriter, r *http.Request) {
	var req createGameReq
	if tryToDecodeOrFail(w, r, &req) != nil {
		return
	}
	log.Printf("Creating a new game named %s with max players %d password %s\n", req.Name, req.MaxPlayersCount, req.Password)
	res, err := a.gameController.NewGame(req.Name, req.MaxPlayersCount, req.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tryToWriteJSONResponse(w, r, res)
}

func (a *API) handleJoinGame(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r)
}

func (a *API) handleLeaveGame(w http.ResponseWriter, r *http.Request) {
	notImplemented(w, r)
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

func notImplemented(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}