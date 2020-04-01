package server

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
)

func (s *Server) mountFileRoutes(r chi.Router) {
	var frontendDir string
	if s.conf.FrontendDir != "" {
		frontendDir = s.conf.FrontendDir
	} else if s.conf.Debug {
		frontendDir = filepath.Join("frontend", "public")
	} else {
		frontendDir = "frontend-build"
	}
	r.Get("/rooms/{roomName}", func(w http.ResponseWriter, r *http.Request) {
		roomName := chi.URLParam(r, "roomName")
		if s.roomManager.Exists(roomName) {
			http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
		} else {
			http.Error(w, http.StatusText(404), 404)
		}
	})
	fs := http.FileServer(http.Dir(frontendDir))
	r.Get("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
