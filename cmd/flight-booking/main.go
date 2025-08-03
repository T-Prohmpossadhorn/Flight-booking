package main

import (
	"github.com/T-Prohmpossadhorn/flight-booking/internal/route"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/flights", route.AddFlightHandler)
	r.GET("/flights/:flight_id", route.GetFlightHandler)
	r.POST("/book", route.BookFlightHandler)
	r.POST("/cancel", route.CancelBookingHandler)
	r.Run(":8080")
}
