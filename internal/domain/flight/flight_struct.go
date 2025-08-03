package flight

import (
	"errors"
	"sync"
	"time"
)

type SeatClass string

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

var errNoSeatAvailable = errors.New("no seat available")
