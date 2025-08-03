package route

import (
	"net/http"
	"time"

	"github.com/T-Prohmpossadhorn/flight-booking/internal/domain/flight"
	"github.com/gin-gonic/gin"
)

func AddFlightHandler(c *gin.Context) {
	var req AddFlightInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid flight data"})
		return
	}
	dep, err := time.Parse("2006-01-02 15:04", req.Departure)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid departure format"})
		return
	}
	arr, err := time.Parse("2006-01-02 15:04", req.Arrival)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid arrival format"})
		return
	}
	fl := flight.InitializeFlight(req.FlightID, req.Origin, req.Destination, req.Aircraft, dep, arr)
	for class, layout := range req.SeatLayout {
		seatLayout := [][]*flight.Seat{}
		for r, row := range layout {
			seatRow := []*flight.Seat{}
			for c, seat := range row {
				seatRow = append(seatRow, &flight.Seat{
					Row:     r + 1,
					Column:  c + 1,
					Special: seat.Special,
				})
			}
			seatLayout = append(seatLayout, seatRow)
		}
		fl.AddSeatClass(flight.SeatClass(class), seatLayout, req.BasePrices[class])
	}
	service.AddFlight(fl)
	c.JSON(http.StatusOK, gin.H{"status": "Flight added"})
}

func GetFlightHandler(c *gin.Context) {
	flightID := c.Param("flight_id")
	fl := service.FindFlightByID(flightID)
	if fl == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flight not found"})
		return
	}
	resp := GetFlightResponse{
		FlightID:    fl.FlightID,
		Origin:      fl.Origin,
		Destination: fl.Destination,
		Departure:   fl.Departure.Format("2006-01-02 15:04"),
		Seats: map[string]struct {
			Total     int     `json:"total"`
			Available int     `json:"available"`
			BasePrice float64 `json:"base_price"`
		}{},
	}
	for class, seats := range fl.Seats {
		total := len(seats)
		available := 0
		for _, s := range seats {
			if !s.IsBooked && s.Special == "" {
				available++
			}
		}
		resp.Seats[string(class)] = struct {
			Total     int     `json:"total"`
			Available int     `json:"available"`
			BasePrice float64 `json:"base_price"`
		}{
			Total:     total,
			Available: available,
			BasePrice: fl.BasePrices[class],
		}
	}
	c.JSON(http.StatusOK, resp)
}

func BookFlightHandler(c *gin.Context) {
	var req BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking request"})
		return
	}
	bookDate, err := time.Parse("2006-01-02", req.BookingDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking_date"})
		return
	}
	bk, err := service.BookSeat(req.PassengerID, req.FlightID, req.SeatClass, bookDate)
	if err != nil {
		// Try to detect upgrade suggestion
		if err.Error() == "no seat available" {
			flightObj := service.FindFlightByID(req.FlightID)
			upgrade := ""
			for _, c := range []string{"Business", "First"} {
				if c != req.SeatClass {
					for _, s := range flightObj.Seats[flight.SeatClass(c)] {
						if !s.IsBooked {
							upgrade = c
							break
						}
					}
				}
				if upgrade != "" {
					break
				}
			}
			if upgrade != "" {
				c.JSON(http.StatusConflict, BookingError{
					Error: "No seats available in " + req.SeatClass + ". Upgrade to " + upgrade + "?",
				})
				return
			}
		}
		c.JSON(http.StatusConflict, BookingError{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, BookingResponse{
		BookingID:   bk.BookingID,
		PassengerID: bk.PassengerID,
		FlightID:    bk.FlightID,
		Seat:        bk.SeatID,
		Price:       bk.Price,
		Status:      "Confirmed",
	})
}

func CancelBookingHandler(c *gin.Context) {
	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cancellation request"})
		return
	}
	bk, err := passengerStore.GetBooking(req.BookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}
	// Save refund before cancellation (since CancelBooking may update price)
	refund := bk.Price * 0.8
	err = service.CancelBooking(req.BookingID, time.Now())
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, CancelResponse{
		BookingID:    req.BookingID,
		Status:       "Cancelled",
		RefundAmount: refund,
	})
}
