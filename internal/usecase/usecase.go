package usecase

import (
	"errors"
	"time"

	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/booking"
	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/flight"
	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/passenger"
)

func (a *bookingMutexAdapter) Lock()   { a.m.Lock() }
func (a *bookingMutexAdapter) Unlock() { a.m.Unlock() }

func (f *bookingFlightAdapter) GetMutex(class string) booking.Mutex {
	return &bookingMutexAdapter{m: f.Flight.GetMutex(class)}
}

func (f *bookingFlightAdapter) GetSeats(class string) []booking.Seat {
	seats := f.Flight.GetSeats(class)
	result := make([]booking.Seat, len(seats))
	for i, s := range seats {
		result[i] = s
	}
	return result
}

func (f *bookingFlightAdapter) GetColumns(class string) int {
	return f.Flight.GetColumns(class)
}

func (f *bookingFlightAdapter) GetRows(class string) int {
	return f.Flight.GetRows(class)
}

func (f *bookingFlightAdapter) GetBasePrice(class string) float64 {
	return f.Flight.GetBasePrice(class)
}

func (f *bookingFlightAdapter) GetDeparture() time.Time {
	return f.Flight.GetDeparture()
}

func NewService(flights []*flight.Flight, passengers passenger.Storage) *Service {
	return &Service{Flights: flights, Passengers: passengers}
}

func (s *Service) AddFlight(f *flight.Flight) {
	s.Flights = append(s.Flights, f)
}

func (s *Service) SearchFlights(origin, destination string, date time.Time) []*flight.Flight {
	var result []*flight.Flight
	for _, f := range s.Flights {
		if f.Origin == origin && f.Destination == destination &&
			sameDay(f.Departure, date) {
			result = append(result, f)
		}
	}
	return result
}

func (s *Service) BookSeat(passengerID, flightID, class string, now time.Time) (*passenger.BookingInfo, error) {
	flightObj := s.findFlightByID(flightID)
	if flightObj == nil {
		return nil, errors.New("flight not found")
	}

	isFrequentFlyer := s.isFrequentFlyer(passengerID)

	bestSeatFunc := func(seats []booking.Seat, col, row int) booking.Seat {
		var flightSeats []*flight.Seat
		for _, s := range seats {
			if fs, ok := s.(*flight.Seat); ok {
				flightSeats = append(flightSeats, fs)
			}
		}
		seat := flight.BestSeat(flightSeats, col, row)
		return seat
	}

	adapter := &bookingFlightAdapter{Flight: flightObj}
	seat, price, err := booking.BookBestSeat(adapter, class, bestSeatFunc, flight.CalculatePrice, isFrequentFlyer)
	if err != nil && errors.Is(err, booking.ErrNoSeatAvailable) {
		upgradeClass, upErr := tryUpgradeClass(flightObj, class)
		if upErr == nil {
			seat, price, err = booking.BookBestSeat(adapter, upgradeClass, bestSeatFunc, flight.CalculatePrice, isFrequentFlyer)
			if err != nil {
				return nil, err
			}
			class = upgradeClass
		} else {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	bookingInfo := &passenger.BookingInfo{
		BookingID:   generateBookingID(),
		PassengerID: passengerID,
		FlightID:    flightID,
		SeatID:      seat.(*flight.Seat).SeatID,
		SeatClass:   class,
		BookedAt:    now,
		Price:       price,
	}
	if err := s.Passengers.SaveBooking(bookingInfo); err != nil {
		return nil, err
	}
	return bookingInfo, nil
}

func (s *Service) CancelBooking(bookingID string, now time.Time) error {
	bookingInfo, err := s.Passengers.GetBooking(bookingID)
	if err != nil {
		return err
	}
	flightObj := s.findFlightByID(bookingInfo.FlightID)
	if flightObj == nil {
		return errors.New("flight not found")
	}

	seatClass := flight.SeatClass(bookingInfo.SeatClass)
	for _, seat := range flightObj.Seats[seatClass] {
		if seat.SeatID == bookingInfo.SeatID {
			seat.IsBooked = false
			break
		}
	}

	return nil
}
