package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/m0ppers/wuerfler/config"
	"github.com/m0ppers/wuerfler/rooms"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Server holds everything the wuerfler server needs
type Server struct {
	router      chi.Router
	log         *log.Logger
	roomManager *rooms.Manager
}

// NewServer returns a new server
func NewServer(conf config.Config) *Server {
	log := logrus.New()
	if conf.Debug {
		log.SetLevel(logrus.DebugLevel)
	}
	server := &Server{
		router:      chi.NewRouter(),
		roomManager: rooms.NewManager(log),
		log:         log,
	}

	r := server.router
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/rooms/{roomName}/websocket", server.websocketHandler)
	r.Group(func(r chi.Router) {
		// set timeout for non classic http calls
		r.Use(middleware.Timeout(60 * time.Second))
		server.mountRestRoutes(r)
		server.mountFileRoutes(r)

	})

	return server
}

// Run runs the server
func (s *Server) Run() {
	prometheus.MustRegister(rooms.RoomsGauge)
	prometheus.MustRegister(ConnectionsGauge)
	s.log.Info("Würfler starting...")
	http.ListenAndServe(":3000", s.router)
}
