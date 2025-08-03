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
	Flights    []*flight.Flight
	Passengers passenger.Storage
}
