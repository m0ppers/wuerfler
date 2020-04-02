package rooms

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	// RoomsGauge keeps track of all the active rooms
	RoomsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "wuerfler_rooms",
		Help: "Active rooms",
	})
)

const (
	// AddRollerErrorRoomNonExistent will be thrown if the room where we tried to add ourselves is non existent
	AddRollerErrorRoomNonExistent = iota
)

const (
	// CachedResults controls the amount of results cached per room that is sent upon reconnect
	CachedResults = 10
	// RoomIdleTime determines when a room is being closed once all members left
	RoomIdleTime = 60 * time.Second
)

// AddRollerError is an error thrown when trying to add rollers to a room
type AddRollerError struct {
	error
	Type int
}

// NewAddRollerError creates a new AddRollerError
func NewAddRollerError(t int) *AddRollerError {
	return &AddRollerError{
		Type: t,
	}
}

func (e *AddRollerError) Error() string {
	return "Room doesn't exist"
}

// UsersUpdateInfo will be sent whenevert something changed regarding to usernames or so on the backend (join, leave, name change)
type UsersUpdateInfo struct {
	Self   string   `json:"self"`
	Others []string `json:"others"`
}

// ProfileUpdateRequest is the input data when somebody tries to change their name
type ProfileUpdateRequest struct {
	OldName string
	NewName string
}

// RollRequest is the request to roll some dices
type RollRequest struct {
	Name  string
	Dices []uint8
}

// RollResult is the result of one dice
type RollResult struct {
	Dice   uint8 `json:"dice"`
	Result uint8 `json:"result"`
}

// RollResults is the result of several dices of a roller
type RollResults struct {
	Name    string       `json:"name"`
	Date    time.Time    `json:"date"`
	Results []RollResult `json:"results"`
}

// Roller is our User object
type Roller struct {
	Name            string
	RollRequestChan chan []uint8
	ProfileUpdate   chan string
	RollResultsChan chan RollResults
	UsersUpdate     chan UsersUpdateInfo
	RemoveSelf      chan struct{}
}

// NewRoller creates a new Roller
func NewRoller(name string) Roller {
	return Roller{
		Name:            name,
		RollRequestChan: make(chan []uint8, 16),
		ProfileUpdate:   make(chan string, 16),
		RollResultsChan: make(chan RollResults, 16),
		UsersUpdate:     make(chan UsersUpdateInfo, 16),
		RemoveSelf:      make(chan struct{}),
	}
}

// Room holds everything room related
type Room struct {
	name      string
	addRoller chan Roller
}

// NewRoom creates a new room
func NewRoom(name string) Room {
	return Room{
		name:      name,
		addRoller: make(chan Roller, 16),
	}
}

// Manager manages rooms
type Manager struct {
	log *log.Logger

	mutex sync.RWMutex
	rooms map[string]Room
}

// NewManager creates a new manager
func NewManager(log *log.Logger) *Manager {
	return &Manager{
		log:   log,
		rooms: make(map[string]Room, 0),
	}
}

func generateUniqueName(existing []string) string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	// safeguard
	for i := 0; i < 100; i++ {
		name := nameGenerator.Generate()

		found := false
		for _, current := range existing {
			if current == name {
				found = true
				break
			}
		}
		if !found {
			return name
		}
	}

	var buffer bytes.Buffer
	for {
		buffer.WriteString("INVALID")

		name := buffer.String()
		found := false
		for _, current := range existing {
			if current == name {
				found = true
				break
			}
		}
		if !found {
			return name
		}
	}
}

func makeUniqueName(name string, names []string) string {
	if name == "" {
		name = generateUniqueName(names)
	}

	baseName := name
	i := 1
	for {
		found := false
		for _, current := range names {
			if current == name {
				found = true
				break
			}
		}

		if !found {
			break
		}
		name = fmt.Sprintf("%s-%d", baseName, i)
		i++
	}
	return name
}

func runRoller(roller *Roller, log *logrus.Entry, removeRoller chan<- string, roll chan<- RollResults, profileUpdate chan<- ProfileUpdateRequest) {
	for {
		select {
		case <-roller.RemoveSelf:
			log.Debugf("Scheduling removal of %s", roller.Name)
			removeRoller <- roller.Name
			return
		case dices := <-roller.RollRequestChan:
			rand.Seed(time.Now().UTC().UnixNano())
			results := make([]RollResult, 0)
			for _, dice := range dices {
				if dice <= 1 {
					continue
				}
				results = append(results, RollResult{
					Dice:   dice,
					Result: uint8(1 + rand.Intn(int(dice))),
				})
			}
			roll <- RollResults{
				Name:    roller.Name,
				Results: results,
				Date:    time.Now(),
			}
		case newName := <-roller.ProfileUpdate:
			profileUpdate <- ProfileUpdateRequest{
				OldName: roller.Name,
				NewName: newName,
			}
		}
	}
}

func addRoller(log *logrus.Entry, rollers []*Roller, roller Roller) []*Roller {
	rollerNames := make([]string, 0, len(rollers))
	for _, other := range rollers {
		rollerNames = append(rollerNames, other.Name)
	}
	name := makeUniqueName(roller.Name, rollerNames)

	roller.Name = name
	rollers = append(rollers, &roller)
	log.Infof("Added user `%s`. New roller count: %d", name, len(rollers))

	sendUserUpdates(log, rollers)
	return rollers
}

func sendUserUpdates(log *logrus.Entry, rollers []*Roller) {
	if len(rollers) == 0 {
		return
	}

	for _, member := range rollers {
		others := make([]string, 0, len(rollers)-1)
		for _, other := range rollers {
			if other.Name != member.Name {
				others = append(others, other.Name)
			}
		}
		log.Debugf("Sending friend list. Roller: %s, Others: %s", member.Name, strings.Join(others, ", "))
		usersUpdate := UsersUpdateInfo{
			Self:   member.Name,
			Others: others,
		}
		member.UsersUpdate <- usersUpdate
	}
}

func removeRoller(log *logrus.Entry, rollers []*Roller, name string) []*Roller {
	log.Debugf("Removing %s", name)
	found := false
	for i, roller := range rollers {
		if roller.Name == name {
			rollers = append(rollers[:i], rollers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		names := make([]string, 0, len(rollers))
		for _, roller := range rollers {
			names = append(names, roller.Name)
		}
		log.Errorf("Couldn't find %s in room list?! in room: %s", name, strings.Join(names, ", "))
	}

	sendUserUpdates(log, rollers)

	log.Debugf("Removed %s", name)
	return rollers
}

func (m *Manager) runRoom(log *logrus.Entry, room Room) {
	log.Infof("Room `%s` started", room.name)
	RoomsGauge.Inc()
	defer func() {
		func() {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			delete(m.rooms, room.name)
		}()
		log.Infof("Room `%s` ended", room.name)
		RoomsGauge.Dec()
	}()
	rollers := make([]*Roller, 0)

	removeRollerChan := make(chan string, 4)
	roll := make(chan RollResults, 16)
	profileUpdate := make(chan ProfileUpdateRequest, 16)

	t := time.NewTimer(RoomIdleTime)

	lastRolls := make([]RollResults, 0, CachedResults)
	for {
		select {
		case roller := <-room.addRoller:
			rollers = addRoller(log, rollers, roller)
			l := len(rollers)
			// must be ptr because goroutine is searching via name and name might change due to profile update
			rollerPtr := rollers[l-1]

			for _, lastRoll := range lastRolls {
				rollerPtr.RollResultsChan <- lastRoll
			}

			// need to stop room end timer if this is the first user
			if l == 1 && !t.Stop() {
				<-t.C
			}
			go runRoller(rollerPtr, log, removeRollerChan, roll, profileUpdate)
		case name := <-removeRollerChan:
			rollers = removeRoller(log, rollers, name)
			if len(rollers) == 0 {
				t.Reset(RoomIdleTime)
			}
		case profileUpdateRequest := <-profileUpdate:
			var l int
			if len(rollers) <= 1 {
				l = 0
			} else {
				l = len(rollers) - 1
			}
			var r *Roller
			others := make([]string, 0, l)
			for _, roller := range rollers {
				if roller.Name == profileUpdateRequest.OldName {
					r = roller
				} else {
					others = append(others, roller.Name)
				}
			}
			if r == nil {
				log.Errorf("Couldn't find roller %s in room", profileUpdateRequest.OldName)
				continue
			}
			r.Name = makeUniqueName(profileUpdateRequest.NewName, others)
			sendUserUpdates(log, rollers)
		case r := <-roll:
			if len(lastRolls) < CachedResults {
				lastRolls = append(lastRolls, r)
			} else {
				for i := 0; i < CachedResults-1; i++ {
					lastRolls[i] = lastRolls[i+1]
				}
				lastRolls[CachedResults-1] = r
			}
			for _, roller := range rollers {
				roller.RollResultsChan <- r
			}
		case <-t.C:
			return
		}

	}
}

// CreateRoom creates a new room and returns the new name
func (m *Manager) CreateRoom(name string) (string, error) {
	for i := 0; i < 100; i++ {
		roomNames := func() []string {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			roomNames := make([]string, 0, len(m.rooms))
			for k := range m.rooms {
				roomNames = append(roomNames, k)
			}
			return roomNames
		}()
		roomName := makeUniqueName(name, roomNames)
		r := NewRoom(roomName)
		ok := func() bool {
			m.mutex.Lock()
			defer m.mutex.Unlock()

			_, exists := m.rooms[roomName]
			if exists {
				return false
			}
			m.rooms[roomName] = r
			return true
		}()
		if ok {
			log := m.log.WithField("room", roomName)
			go m.runRoom(log, r)
			return roomName, nil
		}
	}
	return "", errors.New("Couldn't find a unique roomname")
}

// Exists checks if a roomName exists
func (m *Manager) Exists(roomName string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, ok := m.rooms[roomName]
	return ok
}

// AddRoller adds a new roller to a room
func (m *Manager) AddRoller(roomName string, name string) (Roller, error) {
	room, ok := func() (Room, bool) {
		m.mutex.RLock()
		defer m.mutex.RUnlock()

		room, ok := m.rooms[roomName]
		return room, ok
	}()

	if !ok {
		return Roller{}, NewAddRollerError(AddRollerErrorRoomNonExistent)
	}
	roller := NewRoller(name)
	room.addRoller <- roller

	return roller, nil
}
