package handlers

import (
	_ "context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"luxsuv-backend/data"
	"luxsuv-backend/repository"
	"net/http"
	"strconv"
)

func SetupRiderRouter(repo *repository.BookingRepository) *chi.Mux { // Changed to *chi.Router
	r := chi.NewRouter()

	// Public endpoints for riders
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"message": "Hello world"})
	})
	r.Post("/book-ride", createBookRide(repo))
	r.Put("/book-ride/{id}", updateBookRide(repo))
	r.Get("/book-rides", listBookRidesByEmail(repo))

	return r
}

func createBookRide(repo *repository.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var ride data.BookRide
		if err := json.NewDecoder(r.Body).Decode(&ride); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		if err := validateBookRide(&ride); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}

		id, err := repo.CreateBookRide(ctx, &ride)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to create ride booking: %w", err))
			return
		}

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"message": "Ride booking created successfully",
			"id":      id,
		})
	}
}

func updateBookRide(repo *repository.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("invalid booking ID: %w", err))
			return
		}

		var ride data.BookRide
		if err := json.NewDecoder(r.Body).Decode(&ride); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}
		ride.ID = id

		if err := validateBookRide(&ride); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}

		if err := repo.UpdateBookRide(ctx, &ride); err != nil {
			if err == pgx.ErrNoRows {
				respondError(w, http.StatusNotFound, fmt.Errorf("ride booking not found: %d", id))
				return
			}
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to update ride booking: %w", err))
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "Ride booking updated successfully"})
	}
}

func listBookRidesByEmail(repo *repository.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		email := r.URL.Query().Get("email")
		if email == "" {
			respondError(w, http.StatusBadRequest, fmt.Errorf("email query parameter is required"))
			return
		}

		rides, err := repo.ListBookRidesByEmail(ctx, email)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to list ride bookings: %w", err))
			return
		}

		respondJSON(w, http.StatusOK, rides)
	}
}

func validateBookRide(ride *data.BookRide) error {
	if ride.YourName == "" {
		return fmt.Errorf("your_name is required")
	}
	if ride.Email == "" {
		return fmt.Errorf("email is required")
	}
	if ride.PhoneNumber == "" {
		return fmt.Errorf("phone_number is required")
	}
	if ride.RideType != "hourly" && ride.RideType != "per_ride" {
		return fmt.Errorf("rideType must be 'hourly' or 'per_ride', got %s", ride.RideType)
	}
	if ride.PickupLocation == "" {
		return fmt.Errorf("pickup_location is required")
	}
	if ride.DropoffLocation == "" {
		return fmt.Errorf("dropoff_location is required")
	}
	if ride.Date == "" {
		return fmt.Errorf("date is required")
	}
	if ride.Time == "" {
		return fmt.Errorf("time is required")
	}
	if ride.NumberOfPassengers <= 0 {
		return fmt.Errorf("number_of_passengers must be positive")
	}
	return nil
}
