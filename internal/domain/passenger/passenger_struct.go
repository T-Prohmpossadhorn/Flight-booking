package passenger

import "time"

type BookingInfo struct {
	BookingID   string
	PassengerID string
	FlightID    string
	SeatID      string
	SeatClass   string
	BookedAt    time.Time
	Price       float64
}

type Storage interface {
	SaveBooking(info *BookingInfo) error
	GetBooking(bookingID string) (*BookingInfo, error)
	ListBookingsByPassenger(passengerID string) ([]*BookingInfo, error)
	ListBookingsByFlight(flightID string) ([]*BookingInfo, error)
}

type InMemoryStorage struct {
	bookings map[string]*BookingInfo
}
