package passenger

import (
	"sync"
	"testing"
	"time"
)

func TestInMemoryStorage(t *testing.T) {
	t.Run("SaveAndGetBooking", func(t *testing.T) {
		storage := NewInMemoryStorage()
		booking := &BookingInfo{
			BookingID:   "B001",
			PassengerID: "P001",
			FlightID:    "F001",
			SeatID:      "1A",
			SeatClass:   "Economy",
			BookedAt:    time.Now(),
			Price:       1234.56,
		}
		err := storage.SaveBooking(booking)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			got, err := storage.GetBooking("B001")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got.BookingID != "B001" {
				t.Errorf("unexpected booking info: %+v", got)
			}
		}()
		go func() {
			defer wg.Done()
			got, err := storage.GetBooking("B001")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got.BookingID != "B001" {
				t.Errorf("unexpected booking info: %+v", got)
			}
		}()
		wg.Wait()
	})

	t.Run("GetBooking_NotFound", func(t *testing.T) {
		storage := NewInMemoryStorage()
		_, err := storage.GetBooking("not-exist")
		if err == nil {
			t.Fatal("expected error for not found booking")
		}
		if err != ErrBookingNotFound {
			t.Errorf("expected ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("ListBookingsByPassenger", func(t *testing.T) {
		storage := NewInMemoryStorage()
		b1 := &BookingInfo{BookingID: "B1", PassengerID: "P1"}
		b2 := &BookingInfo{BookingID: "B2", PassengerID: "P1"}
		b3 := &BookingInfo{BookingID: "B3", PassengerID: "P2"}
		storage.SaveBooking(b1)
		storage.SaveBooking(b2)
		storage.SaveBooking(b3)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			list, err := storage.ListBookingsByPassenger("P1")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(list) != 2 {
				t.Errorf("expected 2 bookings, got %d", len(list))
			}
		}()
		go func() {
			defer wg.Done()
			list, err := storage.ListBookingsByPassenger("P2")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(list) != 1 {
				t.Errorf("expected 1 booking, got %d", len(list))
			}
		}()
		wg.Wait()
	})

	t.Run("ListBookingsByFlight", func(t *testing.T) {
		storage := NewInMemoryStorage()
		b1 := &BookingInfo{BookingID: "B1", FlightID: "F1"}
		b2 := &BookingInfo{BookingID: "B2", FlightID: "F1"}
		b3 := &BookingInfo{BookingID: "B3", FlightID: "F2"}
		storage.SaveBooking(b1)
		storage.SaveBooking(b2)
		storage.SaveBooking(b3)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			list, err := storage.ListBookingsByFlight("F1")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(list) != 2 {
				t.Errorf("expected 2 bookings, got %d", len(list))
			}
		}()
		go func() {
			defer wg.Done()
			list, err := storage.ListBookingsByFlight("F2")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(list) != 1 {
				t.Errorf("expected 1 booking, got %d", len(list))
			}
		}()
		wg.Wait()
	})
}

func TestNotFoundError_Error(t *testing.T) {
	t.Run("ErrorMessage", func(t *testing.T) {
		err := NewNotFoundError("not found")
		if err.Error() != "not found" {
			t.Errorf("expected error message 'not found', got %s", err.Error())
		}
	})
}
