package server

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (s *Server) mountFileRoutes(r chi.Router) {
	r.Get("/rooms/{roomName}", func(w http.ResponseWriter, r *http.Request) {
		roomName := chi.URLParam(r, "roomName")
		if s.roomManager.Exists(roomName) {
			http.ServeFile(w, r, "./frontend/public/index.html")
		} else {
			http.Error(w, http.StatusText(404), 404)
		}
	})
	fs := http.FileServer(http.Dir("./frontend/public"))
	r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
