package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/flight"
)

func (s *Service) findFlightByID(flightID string) *flight.Flight {
	for _, f := range s.Flights {
		if f.FlightID == flightID {
			return f
		}
	}
	return nil
}

func (s *Service) isFrequentFlyer(passengerID string) bool {
	bookings, err := s.Passengers.ListBookingsByPassenger(passengerID)
	if err != nil {
		return false
	}
	// Example: 5 or more bookings = frequent flyer
	return len(bookings) >= 5
}

func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// Now a method on Service, uses priority if set
func (s *Service) tryUpgradeClass(f *flight.Flight, currentClass string) (string, error) {
	classes := s.classList(f)
	for i, c := range classes {
		if c == currentClass && i+1 < len(classes) {
			return classes[i+1], nil
		}
	}
	return "", errors.New("no higher class available")
}

// Helper to get all seat classes for a flight, in priority order if set
func (s *Service) classList(f *flight.Flight) []string {
	if len(s.SeatClassPriority) > 0 {
		var present []string
		for _, c := range s.SeatClassPriority {
			if _, ok := f.Seats[flight.SeatClass(c)]; ok {
				present = append(present, c)
			}
		}
		return present
	}
	var classes []string
	for class := range f.Seats {
		classes = append(classes, string(class))
	}
	return classes
}

// Dummy booking ID generator (replace with real one)
func generateBookingID() string {
	return uuid.New().String()
}
