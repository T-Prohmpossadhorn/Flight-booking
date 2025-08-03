package booking

import (
	"errors"
	"time"
)

var ErrNoSeatAvailable = errors.New("no seat available")

type Seat interface {
	IsBookedSeat() bool
	SetBooked(bool)
	GetSpecial() string
}

type Flight interface {
	GetSeats(seatClass string) []Seat
	GetColumns(seatClass string) int
	GetRows(seatClass string) int
	GetBasePrice(seatClass string) float64
	GetDeparture() time.Time
	GetMutex(seatClass string) Mutex
}

type Mutex interface {
	Lock()
	Unlock()
}
