package usecase

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/flight"
	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/passenger"
)

// --- Mock Passenger Storage ---
type mockPassengerStorage struct {
	bookings map[string]*passenger.BookingInfo
}

func (m *mockPassengerStorage) SaveBooking(b *passenger.BookingInfo) error {
	m.bookings[b.BookingID] = b
	return nil
}
func (m *mockPassengerStorage) GetBooking(id string) (*passenger.BookingInfo, error) {
	b, ok := m.bookings[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return b, nil
}
func (m *mockPassengerStorage) ListBookingsByPassenger(pid string) ([]*passenger.BookingInfo, error) {
	var res []*passenger.BookingInfo
	for _, b := range m.bookings {
		if b.PassengerID == pid {
			res = append(res, b)
		}
	}
	return res, nil
}

func (m *mockPassengerStorage) ListBookingsByFlight(fid string) ([]*passenger.BookingInfo, error) {
	var res []*passenger.BookingInfo
	for _, b := range m.bookings {
		if b.FlightID == fid {
			res = append(res, b)
		}
	}
	return res, nil
}

func TestService_BookAndCancelSeat(t *testing.T) {
	t.Run("BookAndCancelSuccess", func(t *testing.T) {
		// Setup flight
		f := &flight.Flight{
			FlightID:    "F1",
			Origin:      "BKK",
			Destination: "JFK",
			Departure:   time.Now().Add(24 * time.Hour),
			Seats: map[flight.SeatClass][]*flight.Seat{"Economy": {
				{SeatID: "1A", Row: 1, Column: 1, IsBooked: false},
			}},
			Columns:    map[flight.SeatClass]int{"Economy": 1},
			Rows:       map[flight.SeatClass]int{"Economy": 1},
			BasePrices: map[flight.SeatClass]float64{"Economy": 1000},
			Mutex:      map[flight.SeatClass]*sync.Mutex{"Economy": new(sync.Mutex)},
		}
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		svc := NewService([]*flight.Flight{f}, passengerStore)

		// Book
		bk, err := svc.BookSeat("P1", "F1", "Economy", time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bk.SeatID != "1A" {
			t.Errorf("expected seat 1A, got %s", bk.SeatID)
		}
		// Cancel
		err = svc.CancelBooking(bk.BookingID, time.Now())
		if err != nil {
			t.Errorf("unexpected cancel error: %v", err)
		}
		if f.Seats["Economy"][0].IsBooked {
			t.Errorf("seat should be unbooked after cancel")
		}
	})

	t.Run("BookNoFlight", func(t *testing.T) {
		svc := NewService([]*flight.Flight{}, &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}})
		_, err := svc.BookSeat("P1", "F404", "Economy", time.Now())
		if err == nil {
			t.Error("expected error for missing flight")
		}
	})

	t.Run("CancelNoBooking", func(t *testing.T) {
		svc := NewService([]*flight.Flight{}, &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}})
		err := svc.CancelBooking("B404", time.Now())
		if err == nil {
			t.Error("expected error for missing booking")
		}
	})

	t.Run("CancelNoFlight", func(t *testing.T) {
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{
			"B1": {BookingID: "B1", FlightID: "F404", SeatClass: "Economy", SeatID: "1A"},
		}}
		svc := NewService([]*flight.Flight{}, passengerStore)
		err := svc.CancelBooking("B1", time.Now())
		if err == nil {
			t.Error("expected error for missing flight")
		}
	})

	t.Run("BookNoSeatAvailable", func(t *testing.T) {
		f := &flight.Flight{
			FlightID:    "F2",
			Origin:      "BKK",
			Destination: "JFK",
			Departure:   time.Now().Add(24 * time.Hour),
			Seats: map[flight.SeatClass][]*flight.Seat{"Economy": {
				{SeatID: "1A", Row: 1, Column: 1, IsBooked: true},
			}},
			Columns:    map[flight.SeatClass]int{"Economy": 1},
			Rows:       map[flight.SeatClass]int{"Economy": 1},
			BasePrices: map[flight.SeatClass]float64{"Economy": 1000},
			Mutex:      map[flight.SeatClass]*sync.Mutex{"Economy": new(sync.Mutex)},
		}
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		svc := NewService([]*flight.Flight{f}, passengerStore)
		_, err := svc.BookSeat("P1", "F2", "Economy", time.Now())
		if err == nil {
			t.Error("expected error for no seat available")
		}
	})
}

func TestService_AddAndSearchFlights(t *testing.T) {
	t.Run("AddAndSearch", func(t *testing.T) {
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		svc := NewService([]*flight.Flight{}, passengerStore)
		dep := time.Now().Add(24 * time.Hour)
		fl := &flight.Flight{
			FlightID:    "F3",
			Origin:      "BKK",
			Destination: "NRT",
			Departure:   dep,
			Seats:       map[flight.SeatClass][]*flight.Seat{},
			Columns:     map[flight.SeatClass]int{},
			Rows:        map[flight.SeatClass]int{},
			BasePrices:  map[flight.SeatClass]float64{},
			Mutex:       map[flight.SeatClass]*sync.Mutex{"Economy": new(sync.Mutex)},
		}
		svc.AddFlight(fl)
		found := svc.SearchFlights("BKK", "NRT", dep)
		if len(found) != 1 || found[0].FlightID != "F3" {
			t.Errorf("expected to find flight F3")
		}
	})
}

func TestService_isFrequentFlyer(t *testing.T) {
	t.Run("FrequentFlyer", func(t *testing.T) {
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		for i := 0; i < 5; i++ {
			id := "B" + string(rune(int('0')+i))
			passengerStore.bookings[id] = &passenger.BookingInfo{
				BookingID:   id,
				PassengerID: "P1",
			}
		}
		svc := NewService([]*flight.Flight{}, passengerStore)
		if !svc.isFrequentFlyer("P1") {
			t.Error("expected frequent flyer")
		}
	})

	t.Run("NotFrequentFlyer", func(t *testing.T) {
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		svc := NewService([]*flight.Flight{}, passengerStore)
		if svc.isFrequentFlyer("P2") {
			t.Error("expected not frequent flyer")
		}
	})
}

func TestService_HelperFunctions(t *testing.T) {
	t.Run("sameDay true", func(t *testing.T) {
		now := time.Now()
		if !sameDay(now, now) {
			t.Error("expected same day")
		}
	})
	t.Run("sameDay false", func(t *testing.T) {
		a := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		b := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		if sameDay(a, b) {
			t.Error("expected not same day")
		}
	})
	t.Run("findFlightByID found", func(t *testing.T) {
		f := &flight.Flight{FlightID: "X"}
		svc := NewService([]*flight.Flight{f}, &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}})
		if svc.findFlightByID("X") != f {
			t.Error("should find flight")
		}
	})
	t.Run("findFlightByID not found", func(t *testing.T) {
		svc := NewService([]*flight.Flight{}, &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}})
		if svc.findFlightByID("Y") != nil {
			t.Error("should not find flight")
		}
	})
	t.Run("tryUpgradeClass success", func(t *testing.T) {
		f := &flight.Flight{
			Seats: map[flight.SeatClass][]*flight.Seat{
				"Economy":  {},
				"Business": {},
			},
		}
		svc := &Service{}
		svc.SetSeatClassPriority([]string{"Economy", "Business"})
		class, err := svc.tryUpgradeClass(f, "Economy")
		if err != nil || class != "Business" {
			t.Errorf("should upgrade to Business, got %s, err: %v", class, err)
		}
	})
	t.Run("tryUpgradeClass fail", func(t *testing.T) {
		f := &flight.Flight{
			Seats: map[flight.SeatClass][]*flight.Seat{
				"Economy": {},
			},
		}
		svc := &Service{}
		_, err := svc.tryUpgradeClass(f, "Economy")
		if err == nil {
			t.Error("should not upgrade")
		}
	})
	t.Run("classList", func(t *testing.T) {
		f := &flight.Flight{
			Seats: map[flight.SeatClass][]*flight.Seat{
				"Economy":  {},
				"Business": {},
			},
		}
		svc := &Service{}
		classes := svc.classList(f)
		if len(classes) != 2 {
			t.Error("expected 2 classes")
		}
	})
	t.Run("generateBookingID", func(t *testing.T) {
		id := generateBookingID()
		if len(id) == 0 {
			t.Error("should generate id")
		}
	})
}

func TestService_BookSeat_UpgradePath(t *testing.T) {
	t.Run("UpgradeToBusinessWithPriority", func(t *testing.T) {
		f := &flight.Flight{
			FlightID:    "F4",
			Origin:      "BKK",
			Destination: "JFK",
			Departure:   time.Now().Add(24 * time.Hour),
			Seats: map[flight.SeatClass][]*flight.Seat{
				"Economy":  {{SeatID: "1A", Row: 1, Column: 1, IsBooked: true}},
				"Business": {{SeatID: "2A", Row: 1, Column: 1, IsBooked: false}},
				"First":    {{SeatID: "3A", Row: 1, Column: 1, IsBooked: false}},
			},
			Columns:    map[flight.SeatClass]int{"Economy": 1, "Business": 1, "First": 1},
			Rows:       map[flight.SeatClass]int{"Economy": 1, "Business": 1, "First": 1},
			BasePrices: map[flight.SeatClass]float64{"Economy": 1000, "Business": 2000, "First": 3000},
			Mutex: map[flight.SeatClass]*sync.Mutex{
				"Economy":  new(sync.Mutex),
				"Business": new(sync.Mutex),
				"First":    new(sync.Mutex),
			},
		}
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		svc := NewService([]*flight.Flight{f}, passengerStore)
		// Set custom priority: Economy -> First -> Business
		svc.SetSeatClassPriority([]string{"Economy", "First", "Business"})
		bk, err := svc.BookSeat("P1", "F4", "Economy", time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bk.SeatClass != "First" {
			t.Errorf("expected upgrade to First, got %s", bk.SeatClass)
		}
	})

	t.Run("UpgradeToBusinessDefaultOrder", func(t *testing.T) {
		f := &flight.Flight{
			FlightID:    "F5",
			Origin:      "BKK",
			Destination: "JFK",
			Departure:   time.Now().Add(24 * time.Hour),
			Seats: map[flight.SeatClass][]*flight.Seat{
				"Economy":  {{SeatID: "1A", Row: 1, Column: 1, IsBooked: true}},
				"Business": {{SeatID: "2A", Row: 1, Column: 1, IsBooked: false}},
			},
			Columns:    map[flight.SeatClass]int{"Economy": 1, "Business": 1},
			Rows:       map[flight.SeatClass]int{"Economy": 1, "Business": 1},
			BasePrices: map[flight.SeatClass]float64{"Economy": 1000, "Business": 2000},
			Mutex: map[flight.SeatClass]*sync.Mutex{
				"Economy":  new(sync.Mutex),
				"Business": new(sync.Mutex),
			},
		}
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		svc := NewService([]*flight.Flight{f}, passengerStore)
		// No priority set, should upgrade to Business (default order)
		bk, err := svc.BookSeat("P1", "F5", "Economy", time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bk.SeatClass != "Business" {
			t.Errorf("expected upgrade to Business, got %s", bk.SeatClass)
		}
	})

	t.Run("NoUpgradeAvailableWithPriority", func(t *testing.T) {
		f := &flight.Flight{
			FlightID:    "F6",
			Origin:      "BKK",
			Destination: "JFK",
			Departure:   time.Now().Add(24 * time.Hour),
			Seats: map[flight.SeatClass][]*flight.Seat{
				"Economy": {{SeatID: "1A", Row: 1, Column: 1, IsBooked: true}},
			},
			Columns:    map[flight.SeatClass]int{"Economy": 1},
			Rows:       map[flight.SeatClass]int{"Economy": 1},
			BasePrices: map[flight.SeatClass]float64{"Economy": 1000},
			Mutex:      map[flight.SeatClass]*sync.Mutex{"Economy": new(sync.Mutex)},
		}
		passengerStore := &mockPassengerStorage{bookings: map[string]*passenger.BookingInfo{}}
		svc := NewService([]*flight.Flight{f}, passengerStore)
		svc.SetSeatClassPriority([]string{"Economy", "Business", "First"})
		_, err := svc.BookSeat("P1", "F6", "Economy", time.Now())
		if err == nil {
			t.Error("expected error for no seat available and no upgrade")
		}
	})
}
