package usecase

import (
	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/flight"
	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/passenger"
)

// --- Mutex Adapter ---
type bookingMutexAdapter struct {
	m flight.MutexInterface
}

// --- Flight Adapter ---
type bookingFlightAdapter struct {
	*flight.Flight
}

type Service struct {
	Flights           []*flight.Flight
	Passengers        passenger.Storage
	SeatClassPriority []string // Highest to lowest, e.g. ["First", "Business", "Economy"]
}

// SetSeatClassPriority sets the seat class upgrade/search order.
func (s *Service) SetSeatClassPriority(priority []string) {
	s.SeatClassPriority = priority
}

// FindFlightByID is an exported wrapper for findFlightByID.
func (s *Service) FindFlightByID(flightID string) *flight.Flight {
	return s.findFlightByID(flightID)
}
