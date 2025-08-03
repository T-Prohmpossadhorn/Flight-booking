package flight

import (
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
}

func TestBookBestSeat(t *testing.T) {
	t.Run("BookBestSeatSuccess", func(t *testing.T) {
		flight := InitializeFlight("FL126", "BKK", "CDG", "Boeing 787", time.Now(), time.Now().Add(12*time.Hour))
		layout := [][]*Seat{
			{&Seat{}, &Seat{}, &Seat{}},
		}
		flight.AddSeatClass("First", layout, 5000.0)

		seat, price, err := flight.BookBestSeat("First", bestSeat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !seat.IsBooked {
			t.Errorf("seat should be booked")
		}
		if price <= 0 {
			t.Errorf("price should be positive, got %f", price)
		}

		for i := 0; i < 2; i++ {
			_, price, err := flight.BookBestSeat("First", bestSeat)
			if err != nil {
				break
			}
			if price <= 0 {
				t.Errorf("price should be positive, got %f", price)
			}
		}

		_, _, err = flight.BookBestSeat("First", bestSeat)
		if err == nil {
			t.Errorf("expected error when booking with no seats available")
		}
	})

	t.Run("BookBestSeatNoSeats", func(t *testing.T) {
		flight := InitializeFlight("FL201", "BKK", "HND", "Boeing 737", time.Now(), time.Now().Add(6*time.Hour))
		_, _, err := flight.BookBestSeat("Economy", bestSeat)
		if err == nil {
			t.Errorf("expected error when booking with no seat class")
		}
	})

	t.Run("BookBestSeatAllSpecial", func(t *testing.T) {
		flight := InitializeFlight("FL202", "BKK", "ICN", "Boeing 737", time.Now(), time.Now().Add(6*time.Hour))
		layout := [][]*Seat{
			{{Special: "Blocked"}},
		}
		flight.AddSeatClass("Economy", layout, 800.0)
		_, _, err := flight.BookBestSeat("Economy", bestSeat)
		if err == nil {
			t.Errorf("expected error when all seats are special")
		}
	})

	t.Run("ConcurrentBookBestSeat", func(t *testing.T) {
		flight := InitializeFlight("FL205", "BKK", "LAX", "Boeing 777", time.Now(), time.Now().Add(15*time.Hour))
		layout := [][]*Seat{
			{&Seat{}, &Seat{}, &Seat{}},
		}
		flight.AddSeatClass("Business", layout, 3000.0)

		done := make(chan struct{})
		for i := 0; i < 3; i++ {
			go func() {
				_, price, _ := flight.BookBestSeat("Business", bestSeat)
				if price <= 0 {
					t.Errorf("price should be positive, got %f", price)
				}
				done <- struct{}{}
			}()
		}
		timeout := time.After(2 * time.Second)
		for i := 0; i < 3; i++ {
			select {
			case <-done:
			case <-timeout:
				t.Fatal("timeout waiting for goroutines")
			}
		}
		for _, seat := range flight.Seats["Business"] {
			if !seat.IsBooked {
				t.Errorf("expected all seats to be booked")
			}
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
		if bestSeat([]*Seat{}, 3, 3) != nil {
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
		seat := bestSeat(seats, 3, 3)
		if seat.Column != 1 && seat.Column != 3 {
			t.Errorf("expected aisle/window seat, got column %d", seat.Column)
		}
	})

	t.Run("PrefersFrontRows", func(t *testing.T) {
		seats := []*Seat{
			{Row: 2, Column: 1},
			{Row: 1, Column: 2},
		}
		seat := bestSeat(seats, 2, 2)
		if seat.Row != 1 {
			t.Errorf("expected front row seat, got row %d", seat.Row)
		}
	})

	t.Run("PrefersLowerColumn", func(t *testing.T) {
		seats := []*Seat{
			{Row: 1, Column: 2},
			{Row: 1, Column: 1},
		}
		seat := bestSeat(seats, 2, 2)
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
		price := calculatePrice(base, departure, bookingDate, 0)
		expected := base * 0.9
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})

	t.Run("LessThanOrEqual7Days", func(t *testing.T) {
		dep := time.Now().Add(5 * 24 * time.Hour)
		price := calculatePrice(base, dep, bookingDate, 0)
		expected := base * 1.2
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})

	t.Run("Between8And30Days", func(t *testing.T) {
		dep := time.Now().Add(15 * 24 * time.Hour)
		price := calculatePrice(base, dep, bookingDate, 0)
		expected := base
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})

	t.Run("WithBookedRatio", func(t *testing.T) {
		price := calculatePrice(base, departure, bookingDate, 0.5)
		expected := base * 0.9 * 1.5
		if price != expected {
			t.Errorf("expected %.2f, got %.2f", expected, price)
		}
	})
}
