package flight

import (
	"sync"
	"testing"
	"time"
)

func TestInitializeFlight(t *testing.T) {
	t.Run("BasicInitialization", func(t *testing.T) {
		departure := time.Now()
		arrival := departure.Add(2 * time.Hour)
		flight := InitializeFlight("FL123", "BKK", "JFK", "Boeing 777", departure, arrival)

		if flight.FlightID != "FL123" {
			t.Errorf("expected FlightID FL123, got %s", flight.FlightID)
		}
		if flight.Origin != "BKK" || flight.Destination != "JFK" {
			t.Errorf("unexpected origin or destination")
		}
		if flight.Aircraft != "Boeing 777" {
			t.Errorf("unexpected aircraft")
		}
		if !flight.Departure.Equal(departure) || !flight.Arrival.Equal(arrival) {
			t.Errorf("unexpected departure or arrival time")
		}
	})
}

func TestAddSeatClass(t *testing.T) {
	t.Run("AddAndCheckSeatClass", func(t *testing.T) {
		flight := InitializeFlight("FL124", "BKK", "LHR", "Airbus A380", time.Now(), time.Now().Add(10*time.Hour))
		layout := [][]*Seat{
			{&Seat{}, &Seat{}, &Seat{}},
			{&Seat{}, &Seat{}, &Seat{}},
		}
		flight.AddSeatClass("Economy", layout, 1000.0)

		if len(flight.Seats["Economy"]) != 6 {
			t.Errorf("expected 6 seats, got %d", len(flight.Seats["Economy"]))
		}
		if flight.BasePrices["Economy"] != 1000.0 {
			t.Errorf("expected base price 1000.0, got %f", flight.BasePrices["Economy"])
		}
		if flight.Columns["Economy"] != 2 || flight.Rows["Economy"] != 3 {
			t.Errorf("unexpected columns or rows")
		}
	})

	t.Run("AddSeatClassTwice", func(t *testing.T) {
		flight := InitializeFlight("FL200", "BKK", "SIN", "Boeing 737", time.Now(), time.Now().Add(2*time.Hour))
		layout := [][]*Seat{
			{&Seat{}},
		}
		flight.AddSeatClass("Economy", layout, 500.0)
		flight.AddSeatClass("Economy", layout, 999.0)
		if flight.BasePrices["Economy"] != 500.0 {
			t.Errorf("expected base price to remain 500.0, got %f", flight.BasePrices["Economy"])
		}
	})

	t.Run("AddSeatClassEmptyLayout", func(t *testing.T) {
		flight := InitializeFlight("FL300", "BKK", "SIN", "Boeing 737", time.Now(), time.Now().Add(2*time.Hour))
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic on empty layout")
			}
		}()
		flight.AddSeatClass("Economy", [][]*Seat{}, 500.0)
	})
}

func TestGetAvailableSeats(t *testing.T) {
	t.Run("AvailableSeatsAfterBooking", func(t *testing.T) {
		flight := InitializeFlight("FL125", "BKK", "SYD", "Boeing 747", time.Now(), time.Now().Add(9*time.Hour))
		layout := [][]*Seat{
			{&Seat{}, &Seat{}},
			{&Seat{}, &Seat{}},
		}
		flight.AddSeatClass("Business", layout, 2000.0)

		available := flight.getAvailableSeats("Business")
		if len(available) != 4 {
			t.Errorf("expected 4 available seats, got %d", len(available))
		}

		available[0].IsBooked = true
		available2 := flight.getAvailableSeats("Business")
		if len(available2) != 3 {
			t.Errorf("expected 3 available seats after booking, got %d", len(available2))
		}
	})

	t.Run("GetAvailableSeatsEmptyClass", func(t *testing.T) {
		flight := InitializeFlight("FL203", "BKK", "DXB", "Boeing 737", time.Now(), time.Now().Add(7*time.Hour))
		available := flight.getAvailableSeats("Economy")
		if len(available) != 0 {
			t.Errorf("expected 0 available seats, got %d", len(available))
		}
	})

	t.Run("AllSeatsSpecialOrBooked", func(t *testing.T) {
		flight := InitializeFlight("FL204", "BKK", "SYD", "Boeing 737", time.Now(), time.Now().Add(9*time.Hour))
		layout := [][]*Seat{
			{{Special: "Blocked"}, {Special: "Blocked"}},
		}
		flight.AddSeatClass("Economy", layout, 1000.0)
		available := flight.getAvailableSeats("Economy")
		if len(available) != 0 {
			t.Errorf("expected 0 available seats, got %d", len(available))
		}
	})
}

func TestSeatIDFormat(t *testing.T) {
	t.Run("SeatIDNotEmpty", func(t *testing.T) {
		flight := InitializeFlight("FL204", "BKK", "SYD", "Boeing 737", time.Now(), time.Now().Add(9*time.Hour))
		layout := [][]*Seat{
			{&Seat{}, &Seat{}},
			{&Seat{}, &Seat{}},
		}
		flight.AddSeatClass("Economy", layout, 1000.0)
		for _, seat := range flight.Seats["Economy"] {
			if seat.SeatID == "" {
				t.Errorf("seat ID should not be empty")
			}
		}
	})
}

func TestBestSeat(t *testing.T) {
	t.Run("EmptySeats", func(t *testing.T) {
		if BestSeat([]*Seat{}, 3, 3) != nil {
			t.Errorf("expected nil for empty seat slice")
		}
	})

	t.Run("PrefersAisleOrWindow", func(t *testing.T) {
		seats := []*Seat{
			{Row: 2, Column: 2},
			{Row: 1, Column: 1}, // window
			{Row: 1, Column: 3}, // window
			{Row: 1, Column: 2},
		}
		seat := BestSeat(seats, 3, 3)
		if seat.Column != 1 && seat.Column != 3 {
			t.Errorf("expected aisle/window seat, got column %d", seat.Column)
		}
	})

	t.Run("PrefersFrontRows", func(t *testing.T) {
		seats := []*Seat{
			{Row: 2, Column: 1},
			{Row: 1, Column: 2},
		}
		seat := BestSeat(seats, 2, 2)
		if seat.Row != 1 {
			t.Errorf("expected front row seat, got row %d", seat.Row)
		}
	})

	t.Run("PrefersLowerColumn", func(t *testing.T) {
		seats := []*Seat{
			{Row: 1, Column: 2},
			{Row: 1, Column: 1},
		}
		seat := BestSeat(seats, 2, 2)
		if seat.Column != 1 {
			t.Errorf("expected lower column seat, got column %d", seat.Column)
		}
	})
}

func TestCalculatePrice(t *testing.T) {
	base := 1000.0
	departure := time.Now().Add(40 * 24 * time.Hour)
	bookingDate := time.Now()

	t.Run("MoreThan30Days", func(t *testing.T) {
		price := CalculatePrice(base, departure, bookingDate, 0, false)
		expected := base * 0.9
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})

	t.Run("LessThanOrEqual7Days", func(t *testing.T) {
		dep := time.Now().Add(5 * 24 * time.Hour)
		price := CalculatePrice(base, dep, bookingDate, 0, false)
		expected := base * 1.2
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})

	t.Run("Between8And30Days", func(t *testing.T) {
		dep := time.Now().Add(15 * 24 * time.Hour)
		price := CalculatePrice(base, dep, bookingDate, 0, false)
		expected := base
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})

	t.Run("WithBookedRatio", func(t *testing.T) {
		price := CalculatePrice(base, departure, bookingDate, 0.5, false)
		expected := base * 0.9 * 1.5
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})

	t.Run("FrequentFlyerDiscount", func(t *testing.T) {
		price := CalculatePrice(base, departure, bookingDate, 0.2, true)
		expected := base * 0.9 * 1.2 * 0.95
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})
}

func TestSeatInterfaceMethods(t *testing.T) {
	seat := &Seat{
		SeatID:   "A1",
		Row:      1,
		Column:   1,
		Special:  "VIP",
		IsBooked: false,
	}
	if seat.GetSeatID() != "A1" {
		t.Errorf("GetSeatID failed")
	}
	if seat.GetRow() != 1 {
		t.Errorf("GetRow failed")
	}
	if seat.GetColumn() != 1 {
		t.Errorf("GetColumn failed")
	}
	if seat.GetSpecial() != "VIP" {
		t.Errorf("GetSpecial failed")
	}
	if seat.IsBookedSeat() {
		t.Errorf("IsBookedSeat should be false")
	}
	seat.SetBooked(true)
	if !seat.IsBookedSeat() {
		t.Errorf("SetBooked or IsBookedSeat failed")
	}
}

func TestFlightInterfaceMethods(t *testing.T) {
	flight := &Flight{
		FlightID:    "F1",
		Origin:      "BKK",
		Destination: "JFK",
		Departure:   time.Now(),
		Arrival:     time.Now().Add(10 * time.Hour),
		Columns:     map[SeatClass]int{"Economy": 2},
		Rows:        map[SeatClass]int{"Economy": 3},
		Seats: map[SeatClass][]*Seat{
			"Economy": {
				{SeatID: "A1"}, {SeatID: "A2"},
				{SeatID: "B1"}, {SeatID: "B2"},
				{SeatID: "C1"}, {SeatID: "C2"},
			},
		},
		BasePrices: map[SeatClass]float64{"Economy": 1000.0},
		Aircraft:   "Boeing 777",
		Mutex:      map[SeatClass]*sync.Mutex{"Economy": &sync.Mutex{}},
	}

	// GetSeats returns correct interface slice
	seats := flight.GetSeats("Economy")
	if len(seats) != 6 {
		t.Errorf("GetSeats returned wrong number of seats")
	}
	if seats[0].GetSeatID() != "A1" {
		t.Errorf("GetSeats returned wrong seat")
	}
	if flight.GetColumns("Economy") != 2 {
		t.Errorf("GetColumns failed")
	}
	if flight.GetRows("Economy") != 3 {
		t.Errorf("GetRows failed")
	}
	if flight.GetBasePrice("Economy") != 1000.0 {
		t.Errorf("GetBasePrice failed")
	}
	if !flight.GetDeparture().Equal(flight.Departure) {
		t.Errorf("GetDeparture failed")
	}
	if flight.GetMutex("Economy") == nil {
		t.Errorf("GetMutex failed")
	}
	if flight.GetMutex("NonExist") != nil {
		t.Errorf("GetMutex should return nil for missing class")
	}
}

func TestMutexAdapter(t *testing.T) {
	var m sync.Mutex
	adapter := (*MutexAdapter)(&m)
	adapter.Lock()
	adapter.Unlock()
}

func TestSeatInterfaceType(t *testing.T) {
	var s SeatInterface = &Seat{SeatID: "A1"}
	if s.GetSeatID() != "A1" {
		t.Errorf("SeatInterface not working")
	}
}

func TestMutexInterfaceType(t *testing.T) {
	var mtx MutexInterface = new(MutexAdapter)
	mtx.Lock()
	mtx.Unlock()
}
