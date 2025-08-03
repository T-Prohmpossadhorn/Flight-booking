package passenger

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{bookings: make(map[string]*BookingInfo)}
}

func (s *InMemoryStorage) SaveBooking(info *BookingInfo) error {
	s.bookings[info.BookingID] = info
	return nil
}

func (s *InMemoryStorage) GetBooking(bookingID string) (*BookingInfo, error) {
	info, ok := s.bookings[bookingID]
	if !ok {
		return nil, ErrBookingNotFound
	}
	return info, nil
}

func (s *InMemoryStorage) ListBookingsByPassenger(passengerID string) ([]*BookingInfo, error) {
	var result []*BookingInfo
	for _, b := range s.bookings {
		if b.PassengerID == passengerID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (s *InMemoryStorage) ListBookingsByFlight(flightID string) ([]*BookingInfo, error) {
	var result []*BookingInfo
	for _, b := range s.bookings {
		if b.FlightID == flightID {
			result = append(result, b)
		}
	}
	return result, nil
}

var ErrBookingNotFound = NewNotFoundError("booking not found")

func NewNotFoundError(msg string) error {
	return &notFoundError{msg: msg}
}

type notFoundError struct {
	msg string
}

func (e *notFoundError) Error() string { return e.msg }
