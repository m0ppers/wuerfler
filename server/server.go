package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/m0ppers/wuerfler/config"
	"github.com/m0ppers/wuerfler/rooms"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

// Server holds everything the wuerfler server needs
type Server struct {
	conf        config.Config
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
		conf:        conf,
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

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, target,
		// see comments below and consider the codes 308, 302, or 301
		http.StatusFound)
}

// Run runs the server
func (s *Server) Run(ctx context.Context) error {
	var httpsServer *http.Server
	var httpHandler http.Handler
	if s.conf.SecurePort > 0 {
		if s.conf.SecureHostname == "" {
			return errors.New("SECUREHOSTNAME not set")
		}
		var certDir string
		if s.conf.CertDir != "" {
			certDir = s.conf.CertDir
		} else {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				return err
			}
			certDir = filepath.Join(cacheDir, "wuerfler-certs")
		}
		if !strings.Contains(strings.Trim(s.conf.SecureHostname, "."), ".") {
			return errors.New("acme/autocert aha: server name component count invalid")
		}
		m := &autocert.Manager{
			Cache:      autocert.DirCache(certDir),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(s.conf.SecureHostname),
		}
		httpsServer = &http.Server{
			Addr:      fmt.Sprintf(":%d", s.conf.SecurePort),
			TLSConfig: m.TLSConfig(),
			Handler:   s.router,
		}
		httpHandler = m.HTTPHandler(http.HandlerFunc(redirect))
	} else {
		httpHandler = s.router
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.conf.Port),
		Handler: httpHandler,
	}

	errCh := make(chan error)
	if httpsServer != nil {
		go func() {
			s.log.Println("Starting HTTPS server.")
			err := httpsServer.ListenAndServeTLS("", "")
			s.log.Println("Ended HTTPS server.")
			errCh <- fmt.Errorf("httpsServer.ListenAndServeTLS: %v", err)
		}()
	}
	go func() {
		s.log.Println("Starting HTTP server.")
		err := httpServer.ListenAndServe()
		s.log.Println("Ended HTTP server.")
		errCh <- fmt.Errorf("httpServer.ListenAndServe: %v", err)
	}()

	prometheus.MustRegister(rooms.RoomsGauge)
	prometheus.MustRegister(ConnectionsGauge)

	select {
	case <-ctx.Done():
		if httpsServer != nil {
			err := httpsServer.Close()
			if err != nil {
				log.Println("httpsServer.Close:", err)
			}
		}
		err := httpServer.Close()
		if err != nil {
			log.Println("httpServer.Close:", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}
