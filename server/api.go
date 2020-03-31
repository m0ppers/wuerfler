package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

func (s *Server) mountRestRoutes(r chi.Router) {
	r.Post("/api/rooms", s.createRoom)
}

func (s *Server) createRoom(w http.ResponseWriter, req *http.Request) {
	var roomName string
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&roomName)
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		return
	}

	if len(roomName) > 1024 {
		// was erlaube?
		http.Error(w, http.StatusText(400), 400)
		return
	}

	roomName, err = s.roomManager.CreateRoom(roomName)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		s.log.Errorf("Couldn't create room: %v", err)
		return
	}

	json, err := json.Marshal(&roomName)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		s.log.Errorf("Error encoding room json: %v", err)
		return
	}
	w.WriteHeader(201)
	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}
