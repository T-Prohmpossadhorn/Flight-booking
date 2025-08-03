package booking_test

import (
	"testing"
	"time"

	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/booking"
	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/booking/mocks"
	"github.com/stretchr/testify/assert"
)

func dummyBestSeat(seats []booking.Seat, col, row int) booking.Seat {
	if len(seats) == 0 {
		return nil
	}
	return seats[0]
}

func dummyCalcPrice(base float64, departure, bookingDate time.Time, bookedRatio float64, isFrequentFlyer bool) float64 {
	return base + bookedRatio*100
}

func TestBookBestSeat_Success(t *testing.T) {
	mockFlight := new(mocks.Flight)
	mockSeat := new(mocks.Seat)
	mockMutex := new(mocks.Mutex)

	seats := []booking.Seat{mockSeat, mockSeat}

	mockFlight.On("GetMutex", "Economy").Return(mockMutex)
	mockFlight.On("GetSeats", "Economy").Return(seats)
	mockFlight.On("GetColumns", "Economy").Return(2)
	mockFlight.On("GetRows", "Economy").Return(1)
	mockFlight.On("GetBasePrice", "Economy").Return(1000.0)
	mockFlight.On("GetDeparture").Return(time.Now())

	mockMutex.On("Lock").Return()
	mockMutex.On("Unlock").Return()

	mockSeat.On("IsBookedSeat").Return(false).Once()
	mockSeat.On("GetSpecial").Return("")
	mockSeat.On("SetBooked", true).Return()
	mockSeat.On("IsBookedSeat").Return(true)

	seat, price, err := booking.BookBestSeat(mockFlight, "Economy", dummyBestSeat, dummyCalcPrice, false)
	assert.NoError(t, err)
	assert.Equal(t, mockSeat, seat)
	assert.Equal(t, 1100.0, price)
}

func TestBookBestSeat_AllBooked(t *testing.T) {
	mockFlight := new(mocks.Flight)
	mockSeat := new(mocks.Seat)
	mockMutex := new(mocks.Mutex)

	mockFlight.On("GetMutex", "Economy").Return(mockMutex)
	mockFlight.On("GetSeats", "Economy").Return([]booking.Seat{mockSeat, mockSeat})
	mockFlight.On("GetColumns", "Economy").Return(2)
	mockFlight.On("GetRows", "Economy").Return(1)
	mockFlight.On("GetBasePrice", "Economy").Return(1000.0)
	mockFlight.On("GetDeparture").Return(time.Now())

	mockMutex.On("Lock").Return()
	mockMutex.On("Unlock").Return()

	// Both seats are booked
	mockSeat.On("IsBookedSeat").Return(true)
	mockSeat.On("GetSpecial").Return("")

	seat, price, err := booking.BookBestSeat(mockFlight, "Economy", dummyBestSeat, dummyCalcPrice, false)
	assert.ErrorIs(t, err, booking.ErrNoSeatAvailable)
	assert.Nil(t, seat)
	assert.Equal(t, 0.0, price)
}

func TestBookBestSeat_NoSuchClass(t *testing.T) {
	mockFlight := new(mocks.Flight)
	mockFlight.On("GetMutex", "NonExist").Return(nil)

	seat, price, err := booking.BookBestSeat(mockFlight, "NonExist", dummyBestSeat, dummyCalcPrice, false)
	assert.ErrorIs(t, err, booking.ErrNoSeatAvailable)
	assert.Nil(t, seat)
	assert.Equal(t, 0.0, price)
}

func TestBookBestSeat_EmptySeats(t *testing.T) {
	mockFlight := new(mocks.Flight)
	mockMutex := new(mocks.Mutex)

	mockFlight.On("GetMutex", "Economy").Return(mockMutex)
	mockFlight.On("GetSeats", "Economy").Return([]booking.Seat{})
	mockFlight.On("GetColumns", "Economy").Return(1)
	mockFlight.On("GetRows", "Economy").Return(1)
	mockFlight.On("GetBasePrice", "Economy").Return(1000.0)
	mockFlight.On("GetDeparture").Return(time.Now())

	mockMutex.On("Lock").Return()
	mockMutex.On("Unlock").Return()

	seat, price, err := booking.BookBestSeat(mockFlight, "Economy", dummyBestSeat, dummyCalcPrice, false)
	assert.ErrorIs(t, err, booking.ErrNoSeatAvailable)
	assert.Nil(t, seat)
	assert.Equal(t, 0.0, price)
}
