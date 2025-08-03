package flight

import (
	"strconv"
	"sync"
	"time"
)

func InitializeFlight(flightID, origin, destination, aircraft string, departure, arrival time.Time) *Flight {
	return &Flight{
		FlightID:    flightID,
		Origin:      origin,
		Destination: destination,
		Departure:   departure,
		Arrival:     arrival,
		Columns:     make(map[SeatClass]int),
		Rows:        make(map[SeatClass]int),
		Seats:       make(map[SeatClass][]*Seat),
		BasePrices:  make(map[SeatClass]float64),
		Aircraft:    aircraft,
		Mutex:       make(map[SeatClass]*sync.Mutex),
	}
}

func (f *Flight) AddSeatClass(seatClass SeatClass, layout [][]*Seat, basePrice float64) {
	if _, exists := f.Seats[seatClass]; !exists {
		f.Seats[seatClass] = make([]*Seat, 0)
		f.BasePrices[seatClass] = basePrice
		f.Mutex[seatClass] = &sync.Mutex{}
	}

	f.Columns[seatClass] = len(layout)
	f.Rows[seatClass] = len(layout[0])

	for column, rows := range layout {
		for row, seat := range rows {
			seat := Seat{
				SeatID:   string(rune(int('A')+column)) + strconv.Itoa(row+1),
				Row:      row + 1,
				Column:   column + 1,
				Special:  seat.Special,
				IsBooked: false,
			}

			f.Seats[seatClass] = append(f.Seats[seatClass], &seat)
		}
	}
}

func (f *Flight) getAvailableSeats(seatClass SeatClass) []*Seat {
	availableSeats := make([]*Seat, 0)
	for _, seat := range f.Seats[seatClass] {
		if !seat.IsBooked && seat.Special == "" {
			availableSeats = append(availableSeats, seat)
		}
	}
	return availableSeats
}

func (f *Flight) BookBestSeat(seatClass SeatClass, bestSeat func([]*Seat, int, int) *Seat) (*Seat, float64, error) {
	mutex, ok := f.Mutex[seatClass]
	if !ok {
		return nil, 0, errNoSeatAvailable
	}
	mutex.Lock()
	defer mutex.Unlock()

	availableSeats := f.getAvailableSeats(seatClass)
	totalSeats := len(f.Seats[seatClass])
	if len(availableSeats) == 0 || totalSeats == 0 {
		return nil, 0, errNoSeatAvailable
	}

	seat := bestSeat(availableSeats, f.Columns[seatClass], f.Rows[seatClass])
	seat.IsBooked = true

	// Calculate booked ratio
	bookedCount := 0
	for _, s := range f.Seats[seatClass] {
		if s.IsBooked {
			bookedCount++
		}
	}
	bookedRatio := float64(bookedCount) / float64(totalSeats)

	basePrice := f.BasePrices[seatClass]
	price := calculatePrice(basePrice, f.Departure, time.Now(), bookedRatio)

	return seat, price, nil
}
