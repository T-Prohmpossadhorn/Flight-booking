package booking

import (
	"time"
)

func BookBestSeat(f Flight, seatClass string, bestSeat func([]Seat, int, int) Seat, calculatePrice func(base float64, departure, bookingDate time.Time, bookedRatio float64, isFrequentFlyer bool) float64, isFrequentFlyer bool) (Seat, float64, error) {
	mutex := f.GetMutex(seatClass)
	if mutex == nil {
		return nil, 0, ErrNoSeatAvailable
	}
	mutex.Lock()
	defer mutex.Unlock()

	seats := f.GetSeats(seatClass)
	availableSeats := make([]Seat, 0)
	for _, seat := range seats {
		if !seat.IsBookedSeat() && seat.GetSpecial() == "" {
			availableSeats = append(availableSeats, seat)
		}
	}
	totalSeats := len(seats)
	if len(availableSeats) == 0 || totalSeats == 0 {
		return nil, 0, ErrNoSeatAvailable
	}

	seat := bestSeat(availableSeats, f.GetColumns(seatClass), f.GetRows(seatClass))
	seat.SetBooked(true)

	bookedCount := 0
	for _, s := range seats {
		if s.IsBookedSeat() {
			bookedCount++
		}
	}
	bookedRatio := float64(bookedCount) / float64(totalSeats)
	basePrice := f.GetBasePrice(seatClass)
	price := calculatePrice(basePrice, f.GetDeparture(), time.Now(), bookedRatio, isFrequentFlyer)

	return seat, price, nil
}
