package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/m0ppers/wuerfler/rooms"
	"github.com/prometheus/client_golang/prometheus"
)

// ErrorMessage is being sent whenever there is an error
type ErrorMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Message is the general type
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// RollPayload contains the requested dices
type RollPayload struct {
	Dices []uint8 `json:"dices"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	// ConnectionsGauge keeps track of all active connections
	ConnectionsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "wuerfler_connections",
		Help: "Active connections",
	})
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func (s *Server) writeWebsocketError(conn *websocket.Conn, externalErr error, internalErr error) error {
	s.log.Errorf("%v: %v", externalErr, internalErr)
	err := conn.WriteJSON(&ErrorMessage{
		Type:    "error",
		Payload: externalErr.Error(),
	})
	if err != nil {
		s.log.Errorf("Couldn't write error to websocket: %v", err)
	}
	return err
}

func (s *Server) runWebsocketReader(done chan<- struct{}, conn *websocket.Conn, roll chan<- []uint8, profileUpdate chan<- string) {
	defer func() {
		var d struct{}
		done <- d
		s.log.Debug("Reader done")
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	var message Message
	for {
		err := conn.ReadJSON(&message)
		if err != nil {
			s.writeWebsocketError(conn, errors.New("Couldn't read JSON"), err)
			return
		}

		switch message.Type {
		case "roll":
			dices := make([]uint8, 0)
			err = json.Unmarshal(message.Payload, &dices)

			if err != nil {
				s.writeWebsocketError(conn, errors.New("Internal Error"), err)
				return
			}
			for _, dice := range dices {
				switch dice {
				case 4, 6, 8, 10, 12, 20, 100:
				default:
					s.writeWebsocketError(conn, errors.New("Invalid Dices"), nil)
				}

			}
			roll <- dices
		case "profileUpdate":
			var newName string
			err = json.Unmarshal(message.Payload, &newName)

			if err != nil {
				s.writeWebsocketError(conn, errors.New("Internal Error"), err)
				return
			}

			profileUpdate <- newName
		default:
			s.log.Warnf("Unhandled message type %s", message.Type)
		}

	}

}

func (s *Server) writeMessage(conn *websocket.Conn, t string, p interface{}) error {
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	payload, err := json.Marshal(p)

	if err != nil {
		return fmt.Errorf("Couldn't marshal JSON: %v", err)
	}
	message := Message{
		Type:    t,
		Payload: payload,
	}

	if err := conn.WriteJSON(&message); err != nil {
		return fmt.Errorf("Couldn't write JSON: %v", err)
	}
	return nil
}

func (s *Server) runWebsocketWriter(done chan<- struct{}, conn *websocket.Conn, rollInfo <-chan rooms.RollResults, usersUpdateChan <-chan rooms.UsersUpdateInfo) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.log.Debug("Writer Done")
		var d struct{}
		done <- d
	}()
	for {
		select {
		case roll := <-rollInfo:
			if err := s.writeMessage(conn, "roll", &roll); err != nil {
				s.log.Error(err)
				return
			}
		case usersUpdate := <-usersUpdateChan:
			if err := s.writeMessage(conn, "usersupdate", &usersUpdate); err != nil {
				s.log.Error(err)
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.writeWebsocketError(conn, errors.New("Error upgrading to websocket"), err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	ConnectionsGauge.Inc()
	defer func() {
		conn.Close()
		ConnectionsGauge.Dec()
	}()

	var message Message
	err = conn.ReadJSON(&message)
	if err != nil {
		s.writeWebsocketError(conn, errors.New("Couldn't read JSON"), err)
		return
	}

	if message.Type != "join" {
		s.writeWebsocketError(conn, fmt.Errorf("Invalid initial message"), nil)
		return
	}

	var name string
	err = json.Unmarshal(message.Payload, &name)

	if err != nil {
		s.writeWebsocketError(conn, errors.New("Internal Error"), err)
		return
	}

	roomName := chi.URLParam(r, "roomName")
	roller, err := s.roomManager.AddRoller(roomName, name)
	if err != nil {
		s.writeWebsocketError(conn, errors.New("Internal Error"), err)
		return
	}

	done := make(chan struct{})
	go s.runWebsocketReader(done, conn, roller.RollRequestChan, roller.ProfileUpdate)
	go s.runWebsocketWriter(done, conn, roller.RollResultsChan, roller.UsersUpdate)
	<-done

	var remove struct{}
	roller.RemoveSelf <- remove
}
