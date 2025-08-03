package route

// AddFlight endpoint expects a full seat layout with specials
type AddFlightInput struct {
	FlightID    string `json:"flight_id"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Departure   string `json:"departure"` // "YYYY-MM-DD HH:MM"
	Arrival     string `json:"arrival"`   // "YYYY-MM-DD HH:MM"
	Aircraft    string `json:"aircraft"`
	SeatLayout  map[string][][]struct {
		Special string `json:"special"`
	} `json:"seat_layout"` // class -> 2D layout, each seat can have a special
	BasePrices map[string]float64 `json:"base_prices"`
}

// GetFlight returns AddFlightRequest-style response
type GetFlightResponse struct {
	FlightID    string `json:"flight_id"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Departure   string `json:"departure"`
	Seats       map[string]struct {
		Total     int     `json:"total"`
		Available int     `json:"available"`
		BasePrice float64 `json:"base_price"`
	} `json:"seats"`
}

type BookingRequest struct {
	PassengerID string `json:"passenger_id"`
	FlightID    string `json:"flight_id"`
	SeatClass   string `json:"seat_class"`
	BookingDate string `json:"booking_date"` // "YYYY-MM-DD"
}

type BookingResponse struct {
	BookingID   string  `json:"booking_id"`
	PassengerID string  `json:"passenger_id"`
	FlightID    string  `json:"flight_id"`
	Seat        string  `json:"seat"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
}

type BookingError struct {
	Error string `json:"error"`
}

type CancelRequest struct {
	BookingID string `json:"booking_id"`
}
type CancelResponse struct {
	BookingID    string  `json:"booking_id"`
	Status       string  `json:"status"`
	RefundAmount float64 `json:"refund_amount"`
}
