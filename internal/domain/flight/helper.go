package flight

import (
	"sort"
	"time"
)

func BestSeat(seats []*Seat, col, row int) *Seat {
	if len(seats) == 0 {
		return nil
	}
	sort.Slice(seats, func(i, j int) bool {
		isAisleOrWindowI := seats[i].Column == 1 || seats[i].Column == col
		isAisleOrWindowJ := seats[j].Column == 1 || seats[j].Column == col
		if isAisleOrWindowI != isAisleOrWindowJ {
			return isAisleOrWindowI
		}

		if seats[i].Row != seats[j].Row {
			return seats[i].Row < seats[j].Row
		}

		return seats[i].Column < seats[j].Column
	})
	return seats[0]
}

func CalculatePrice(base float64, departure, bookingDate time.Time, bookedRatio float64) float64 {
	daysDiff := int(departure.Sub(bookingDate).Hours() / 24)
	var price = base

	switch {
	case daysDiff > 30:
		price *= 0.9
	case daysDiff <= 7:
		price *= 1.2
	}

	price *= (1 + bookedRatio)

	return price
}
