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

func (f *Flight) GetSeats(seatClass string) []SeatInterface {
	seats := f.Seats[SeatClass(seatClass)]
	result := make([]SeatInterface, len(seats))
	for i, s := range seats {
		result[i] = s
	}
	return result
}

func (f *Flight) GetColumns(seatClass string) int {
	return f.Columns[SeatClass(seatClass)]
}

func (f *Flight) GetRows(seatClass string) int {
	return f.Rows[SeatClass(seatClass)]
}

func (f *Flight) GetBasePrice(seatClass string) float64 {
	return f.BasePrices[SeatClass(seatClass)]
}

func (f *Flight) GetDeparture() time.Time {
	return f.Departure
}

func (f *Flight) GetMutex(seatClass string) MutexInterface {
	m, ok := f.Mutex[SeatClass(seatClass)]
	if !ok {
		return nil
	}
	return (*MutexAdapter)(m)
}

func (s *Seat) GetSeatID() string {
	return s.SeatID
}

func (s *Seat) GetRow() int {
	return s.Row
}

func (s *Seat) GetColumn() int {
	return s.Column
}

func (s *Seat) GetSpecial() string {
	return s.Special
}

func (s *Seat) IsBookedSeat() bool {
	return s.IsBooked
}

func (s *Seat) SetBooked(b bool) {
	s.IsBooked = b
}

func (m *MutexAdapter) Lock() {
	(*sync.Mutex)(m).Lock()
}

func (m *MutexAdapter) Unlock() {
	(*sync.Mutex)(m).Unlock()
}
