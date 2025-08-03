package route

import (
	"sync"

	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/flight"
	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/passenger"
	"github.com/T-Prohmpossadhorn/flight-booking/internal/usecase"
)

type memoryPassengerStorage struct {
	mu       sync.Mutex
	bookings map[string]*passenger.BookingInfo
}

func newMemoryPassengerStorage() *memoryPassengerStorage {
	return &memoryPassengerStorage{bookings: make(map[string]*passenger.BookingInfo)}
}
func (m *memoryPassengerStorage) SaveBooking(b *passenger.BookingInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bookings[b.BookingID] = b
	return nil
}
func (m *memoryPassengerStorage) GetBooking(id string) (*passenger.BookingInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.bookings[id]
	if !ok {
		return nil, passenger.ErrBookingNotFound
	}
	return b, nil
}
func (m *memoryPassengerStorage) ListBookingsByPassenger(pid string) ([]*passenger.BookingInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*passenger.BookingInfo
	for _, b := range m.bookings {
		if b.PassengerID == pid {
			res = append(res, b)
		}
	}
	return res, nil
}
func (m *memoryPassengerStorage) ListBookingsByFlight(fid string) ([]*passenger.BookingInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*passenger.BookingInfo
	for _, b := range m.bookings {
		if b.FlightID == fid {
			res = append(res, b)
		}
	}
	return res, nil
}

var (
	passengerStore = newMemoryPassengerStorage()
	service        = usecase.NewService([]*flight.Flight{}, passengerStore)
)
