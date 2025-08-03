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

func CalculatePrice(base float64, departure, bookingDate time.Time, bookedRatio float64, isFrequentFlyer bool) float64 {
	price := base
	days := int(departure.Sub(bookingDate).Hours() / 24)
	switch {
	case days >= 30:
		price *= 0.9
	case days <= 7:
		price *= 1.2
	}
	price *= (1 + bookedRatio)
	if isFrequentFlyer {
		price *= 0.95 // 5% discount
	}
	return price
}
