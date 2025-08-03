package flight

import (
	"errors"
	"sync"
	"time"
)

type SeatClass string
type MutexAdapter sync.Mutex

var ErrNoSeatAvailable = errors.New("no seat available")

type SeatInterface interface {
	GetSeatID() string
	GetRow() int
	GetColumn() int
	GetSpecial() string
	IsBookedSeat() bool
	SetBooked(bool)
}

type MutexInterface interface {
	Lock()
	Unlock()
}

type Seat struct {
	SeatID   string
	Row      int
	Column   int
	Special  string
	IsBooked bool
}

type Flight struct {
	FlightID    string
	Origin      string
	Destination string
	Departure   time.Time
	Arrival     time.Time
	Columns     map[SeatClass]int
	Rows        map[SeatClass]int
	Seats       map[SeatClass][]*Seat
	BasePrices  map[SeatClass]float64
	Aircraft    string
	Mutex       map[SeatClass]*sync.Mutex
}
