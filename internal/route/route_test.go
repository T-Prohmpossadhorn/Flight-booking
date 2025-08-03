package route

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type SeatLayoutCell struct {
	Special string `json:"special"`
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/flights", AddFlightHandler)
	r.GET("/flights/:flight_id", GetFlightHandler)
	r.POST("/book", BookFlightHandler)
	r.POST("/cancel", CancelBookingHandler)
	return r
}

func TestAddAndGetFlight(t *testing.T) {
	router := setupTestRouter()

	flightReq := AddFlightInput{
		FlightID:    "AB123",
		Origin:      "JFK",
		Destination: "LAX",
		Departure:   "2024-07-10 08:00",
		Arrival:     "2024-07-10 11:00",
		Aircraft:    "Boeing 777",
		SeatLayout: map[string][][]struct {
			Special string `json:"special"`
		}{
			"Economy": {
				{{Special: ""}, {Special: ""}},
				{{Special: ""}, {Special: ""}},
			},
			"Business": {
				{{Special: ""}},
			},
		},
		BasePrices: map[string]float64{
			"Economy":  300,
			"Business": 1000,
		},
	}
	body, _ := json.Marshal(flightReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/flights", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Get flight
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/flights/AB123", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp GetFlightResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "AB123", resp.FlightID)
	assert.Equal(t, "JFK", resp.Origin)
	assert.Equal(t, "LAX", resp.Destination)
	assert.Contains(t, resp.Seats, "Economy")
	assert.Contains(t, resp.Seats, "Business")
}

func TestAddFlight_InvalidInput(t *testing.T) {
	router := setupTestRouter()

	// Bad JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/flights", bytes.NewBuffer([]byte("{bad json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// Bad date format
	flightReq := AddFlightInput{
		FlightID:    "BADDATE",
		Origin:      "JFK",
		Destination: "LAX",
		Departure:   "bad-date",
		Arrival:     "2024-07-10 11:00",
		Aircraft:    "Boeing 777",
		SeatLayout: map[string][][]struct {
			Special string "json:\"special\""
		}{},
		BasePrices: map[string]float64{},
	}
	body, _ := json.Marshal(flightReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/flights", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestGetFlight_NotFound(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/flights/NOTFOUND", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestBookAndCancelFlight(t *testing.T) {
	router := setupTestRouter()

	// Add flight first
	flightReq := AddFlightInput{
		FlightID:    "CD456",
		Origin:      "BKK",
		Destination: "NRT",
		Departure:   "2024-08-01 09:00",
		Arrival:     "2024-08-01 15:00",
		Aircraft:    "Airbus A350",
		SeatLayout: map[string][][]struct {
			Special string `json:"special"`
		}{
			"Economy": {
				{{Special: ""}, {Special: ""}},
			},
		},
		BasePrices: map[string]float64{
			"Economy": 500,
		},
	}
	body, _ := json.Marshal(flightReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/flights", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Book a seat
	bookReq := BookingRequest{
		PassengerID: "P001",
		FlightID:    "CD456",
		SeatClass:   "Economy",
		BookingDate: time.Now().Format("2006-01-02"),
	}
	body, _ = json.Marshal(bookReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var bookResp BookingResponse
	err := json.Unmarshal(w.Body.Bytes(), &bookResp)
	assert.NoError(t, err)
	assert.Equal(t, "P001", bookResp.PassengerID)
	assert.Equal(t, "CD456", bookResp.FlightID)
	assert.Equal(t, "Confirmed", bookResp.Status)
	assert.NotEmpty(t, bookResp.BookingID)

	// Cancel the booking
	cancelReq := CancelRequest{
		BookingID: bookResp.BookingID,
	}
	body, _ = json.Marshal(cancelReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var cancelResp CancelResponse
	err = json.Unmarshal(w.Body.Bytes(), &cancelResp)
	assert.NoError(t, err)
	assert.Equal(t, bookResp.BookingID, cancelResp.BookingID)
	assert.Equal(t, "Cancelled", cancelResp.Status)
	assert.True(t, cancelResp.RefundAmount > 0)
}

func TestBook_InvalidInput(t *testing.T) {
	router := setupTestRouter()

	// Bad JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/book", bytes.NewBuffer([]byte("{bad json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	// Bad date
	bookReq := BookingRequest{
		PassengerID: "P001",
		FlightID:    "CD456",
		SeatClass:   "Economy",
		BookingDate: "bad-date",
	}
	body, _ := json.Marshal(bookReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestBook_FlightNotFound(t *testing.T) {
	router := setupTestRouter()
	bookReq := BookingRequest{
		PassengerID: "P001",
		FlightID:    "NOTFOUND",
		SeatClass:   "Economy",
		BookingDate: time.Now().Format("2006-01-02"),
	}
	body, _ := json.Marshal(bookReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 409, w.Code)
}

func TestCancel_InvalidInput(t *testing.T) {
	router := setupTestRouter()
	// Bad JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer([]byte("{bad json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestCancel_BookingNotFound(t *testing.T) {
	router := setupTestRouter()
	cancelReq := CancelRequest{
		BookingID: "NOTFOUND",
	}
	body, _ := json.Marshal(cancelReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestBookNoSeatAvailable(t *testing.T) {
	router := setupTestRouter()

	// Add flight with 1 seat
	flightReq := AddFlightInput{
		FlightID:    "EF789",
		Origin:      "BKK",
		Destination: "SIN",
		Departure:   "2024-09-01 10:00",
		Arrival:     "2024-09-01 13:00",
		Aircraft:    "Boeing 737",
		SeatLayout: map[string][][]struct {
			Special string `json:"special"`
		}{
			"Economy": {
				{{Special: ""}},
			},
		},
		BasePrices: map[string]float64{
			"Economy": 200,
		},
	}
	body, _ := json.Marshal(flightReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/flights", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Book the only seat
	bookReq := BookingRequest{
		PassengerID: "P001",
		FlightID:    "EF789",
		SeatClass:   "Economy",
		BookingDate: time.Now().Format("2006-01-02"),
	}
	body, _ = json.Marshal(bookReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Try to book again, should fail
	bookReq.PassengerID = "P002"
	body, _ = json.Marshal(bookReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 409, w.Code)
	var errResp BookingError
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, errResp.Error, "no seat available")
}

func TestBook_UpgradeSuggestion(t *testing.T) {
	router := setupTestRouter()

	// Add flight with full Economy, available Business
	flightReq := AddFlightInput{
		FlightID:    "UPG001",
		Origin:      "BKK",
		Destination: "HND",
		Departure:   "2024-10-01 10:00",
		Arrival:     "2024-10-01 18:00",
		Aircraft:    "Boeing 787",
		SeatLayout: map[string][][]struct {
			Special string `json:"special"`
		}{
			"Economy": {
				{{Special: ""}},
			},
			"Business": {
				{{Special: ""}},
			},
		},
		BasePrices: map[string]float64{
			"Economy":  400,
			"Business": 1200,
		},
	}
	body, _ := json.Marshal(flightReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/flights", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Book the only Economy seat
	bookReq := BookingRequest{
		PassengerID: "P001",
		FlightID:    "UPG001",
		SeatClass:   "Economy",
		BookingDate: time.Now().Format("2006-01-02"),
	}
	body, _ = json.Marshal(bookReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Try to book Economy again, should get upgrade suggestion
	bookReq.PassengerID = "P002"
	body, _ = json.Marshal(bookReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/book", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, 409, w.Code)
	var errResp BookingError
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, errResp.Error, "No seats available in Economy. Upgrade to Business?")
}
