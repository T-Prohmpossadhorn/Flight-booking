# Flight Booking API

This project provides a simple flight booking API with endpoints for managing flights, booking seats, and handling cancellations.

## Getting Started

1. **Run the API server**  
   Make sure your server is running on `http://localhost:8080`.

2. **Import the Postman Collection**  
   - Open [Postman](https://www.postman.com/).
   - Click **Import**.
   - Select the file [`postman.json`](./postman.json) from this repository.

3. **Use the Collection**  
   The collection includes the following requests:

   - **Add Flight**  
     `POST /flights`  
     Adds a new flight with seat layout and base prices.

   - **Get Flight**  
     `GET /flights/:flight_id`  
     Retrieves flight details and seat availability.

   - **Book Seat (Economy)**  
     `POST /book`  
     Books a seat in the specified class.

   - **Book Seat (No Seat Available)**  
     `POST /book`  
     Attempts to book when no seats are available (should return an error).

   - **Book Seat (Upgrade Suggestion)**  
     `POST /book`  
     Attempts to book when the requested class is full but an upgrade is possible (should return an upgrade suggestion).

   - **Cancel Booking**  
     `POST /cancel`  
     Cancels a booking.  
     **Note:** Replace `<PASTE_BOOKING_ID_FROM_BOOK_SEAT_RESPONSE>` with the actual `booking_id` returned from a successful booking.

   - **Get Flight Not Found**  
     `GET /flights/NOTFOUND`  
     Attempts to get a non-existent flight (should return 404).

   - **Book Seat Invalid Data**  
     `POST /book`  
     Sends invalid JSON to test error handling.

## Example: Add a Flight

```json
POST /flights
{
  "flight_id": "AB123",
  "origin": "JFK",
  "destination": "LAX",
  "departure": "2024-07-10 08:00",
  "arrival": "2024-07-10 11:00",
  "aircraft": "Boeing 777",
  "seat_layout": {
    "Economy": [
      [ { "special": "" }, { "special": "" } ],
      [ { "special": "" }, { "special": "" } ]
    ],
    "Business": [
      [ { "special": "" } ]
    ]
  },
  "base_prices": {
    "Economy": 300,
    "Business": 1000
  }
}
```

## Notes

- All endpoints expect and return JSON.
- Dates must be in the format `YYYY-MM-DD HH:mm` for flights and `YYYY-MM-DD` for bookings.
- Booking and cancellation responses include status and IDs for further actions.

## The seat classes can be anything!

MIT License
2025 T-Prohmpossadhorn