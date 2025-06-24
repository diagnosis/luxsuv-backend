package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"luxsuv-backend/data"
	"time"
)

type BookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(ctx context.Context, connString string) (*BookingRepository, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection config: %w", err)
	}
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	return &BookingRepository{db: dbpool}, nil
}

func (r *BookingRepository) Close() {
	r.db.Close()
}

func (r *BookingRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *BookingRepository) CreateBookRide(ctx context.Context, bookRide *data.BookRide) (int64, error) {
	if bookRide.RideType != "hourly" && bookRide.RideType != "per_ride" {
		return 0, fmt.Errorf("invalid rideType: must be 'hourly' or 'per_ride', got %s", bookRide.RideType)
	}
	if bookRide.Email == "" {
		return 0, fmt.Errorf("email is required for rider access")
	}
	query := `
		INSERT INTO book_rides (
		                        your_name,
		                        email,
		                        phone_number,
		                        ride_type,
		                        pickup_location,
		                        dropoff_location,
		                        date,
		                        time,
		                        number_of_passengers,
		                        number_of_luggage,
		                        additional_notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	var generatedID int64
	err := r.db.QueryRow(
		ctx,
		query,
		bookRide.YourName,
		bookRide.Email,
		bookRide.PhoneNumber,
		bookRide.RideType,
		bookRide.PickupLocation,
		bookRide.DropoffLocation,
		bookRide.Date,
		bookRide.Time,
		bookRide.NumberOfPassengers,
		bookRide.NumberOfLuggage,
		bookRide.AdditionalNotes,
	).Scan(&generatedID)
	if err != nil {
		return 0, fmt.Errorf("failed to create book ride: %w", err)
	}
	return generatedID, nil
}

func (r *BookingRepository) ListBookRidesByEmail(ctx context.Context, email string) ([]*data.BookRide, error) {
	query := `
		SELECT id,
		       your_name,
		       email,
		       phone_number,
		       ride_type,
		       pickup_location,
		       dropoff_location,
		       date,
		       time,
		       number_of_passengers,
		       number_of_luggage,
		       additional_notes
		FROM book_rides WHERE email = $1`
	rows, err := r.db.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to list book rides by email: %w", err)
	}
	defer rows.Close()
	var rides []*data.BookRide
	for rows.Next() {
		ride := &data.BookRide{}
		err := rows.Scan(
			&ride.ID,
			&ride.YourName,
			&ride.Email,
			&ride.PhoneNumber,
			&ride.RideType,
			&ride.PickupLocation,
			&ride.DropoffLocation,
			&ride.Date,
			&ride.Time,
			&ride.NumberOfPassengers,
			&ride.NumberOfLuggage,
			&ride.AdditionalNotes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book ride: %w", err)
		}
		rides = append(rides, ride)
	}
	return rides, nil
}

func (r *BookingRepository) ListBookRides(ctx context.Context) ([]*data.BookRide, error) {
	query := `
		SELECT id,
		       your_name,
		       email,
		       phone_number,
		       ride_type,
		       pickup_location,
		       dropoff_location,
		       date,
		       time,
		       number_of_passengers,
		       number_of_luggage,
		       additional_notes
		FROM book_rides`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list book rides: %w", err)
	}
	defer rows.Close()
	var rides []*data.BookRide
	for rows.Next() {
		ride := &data.BookRide{}
		err := rows.Scan(
			&ride.ID,
			&ride.YourName,
			&ride.Email,
			&ride.PhoneNumber,
			&ride.RideType,
			&ride.PickupLocation,
			&ride.DropoffLocation,
			&ride.Date,
			&ride.Time,
			&ride.NumberOfPassengers,
			&ride.NumberOfLuggage,
			&ride.AdditionalNotes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book ride: %w", err)
		}
		rides = append(rides, ride)
	}
	return rides, nil
}

func (r *BookingRepository) GetBookRideByID(ctx context.Context, id int64) (*data.BookRide, error) {
	query := `
        SELECT id, your_name, email, phone_number, ride_type, pickup_location, 
               dropoff_location, date, time, number_of_passengers, number_of_luggage, additional_notes
        FROM book_rides WHERE id = $1`
	ride := &data.BookRide{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ride.ID, &ride.YourName, &ride.Email, &ride.PhoneNumber, &ride.RideType,
		&ride.PickupLocation, &ride.DropoffLocation, &ride.Date, &ride.Time,
		&ride.NumberOfPassengers, &ride.NumberOfLuggage, &ride.AdditionalNotes)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get ride booking: %w", err)
	}
	return ride, nil
}

func (r *BookingRepository) UpdateBookRide(ctx context.Context, ride *data.BookRide) error {
	if ride.RideType != "hourly" && ride.RideType != "per_ride" {
		return fmt.Errorf("invalid rideType: must be 'hourly' or 'per_ride', got %s", ride.RideType)
	}
	query := `
        UPDATE book_rides SET 
            your_name = $2, email = $3, phone_number = $4, ride_type = $5, 
            pickup_location = $6, dropoff_location = $7, date = $8, time = $9, 
            number_of_passengers = $10, number_of_luggage = $11, additional_notes = $12
        WHERE id = $1`
	result, err := r.db.Exec(ctx, query,
		ride.ID,
		ride.YourName,
		ride.Email,
		ride.PhoneNumber,
		ride.RideType,
		ride.PickupLocation,
		ride.DropoffLocation,
		ride.Date,
		ride.Time,
		ride.NumberOfPassengers,
		ride.NumberOfLuggage,
		ride.AdditionalNotes)
	if err != nil {
		return fmt.Errorf("failed to update ride booking: %w", err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	// Placeholder for notification
	fmt.Printf("Update notification sent to email: %s, phone: %s\n", ride.Email, ride.PhoneNumber)
	return nil
}

func (r *BookingRepository) DeleteBookRide(ctx context.Context, id int64) error {
	// Fetch ride details before deletion
	ride, err := r.GetBookRideByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("failed to fetch ride for notification: %w", err)
	}

	query := `DELETE FROM book_rides WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete ride booking: %w", err)
	}
	rowsAffected := result.RowsAffected()
	fmt.Printf("Rows affected by delete for ID %d: %d\n", id, rowsAffected)
	if rowsAffected == 0 {
		return pgx.ErrNoRows
	}
	// Notification placeholder
	fmt.Printf("Delete notification sent to email: %s, phone: %s\n", ride.Email, ride.PhoneNumber)
	return nil
}

func (r *BookingRepository) GetUserByCredentials(ctx context.Context, username, password string) (*data.User, error) {
	query := `
        SELECT id, username, password, role, created_at
        FROM users WHERE username = $1`
	user := &data.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, pgx.ErrNoRows // Treat mismatch as invalid credentials
	}
	return user, nil
}
