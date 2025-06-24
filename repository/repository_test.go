package repository

import (
	"context"
	"fmt"
	"luxsuv-backend/data"
	"strings"
	"testing"
	"time"
)

func TestCreateAndGetBookRide(t *testing.T) {
	ctx := context.Background()
	repo, err := NewBookingRepository(ctx, "postgresql://postgres:postgres@localhost:5432/luxsuv?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ride := &data.BookRide{
		YourName:           "John Doe",
		Email:              "john@example.com",
		PhoneNumber:        "123-456-7890",
		RideType:           "hourly",
		PickupLocation:     "123 Main St",
		DropoffLocation:    "456 Elm St",
		Date:               "2025-06-23",
		Time:               "14:30",
		NumberOfPassengers: 2,
		NumberOfLuggage:    1,
		AdditionalNotes:    "Please arrive early",
	}
	id, err := repo.CreateBookRide(ctx, ride)
	if err != nil {
		t.Fatalf("Failed to create ride: %v", err)
	}
	fmt.Printf("Created ride with ID: %d\n", id)

	retrievedRide, err := repo.GetBookRideByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get ride: %v", err)
	}
	if retrievedRide.ID != id {
		t.Errorf("Expected ID %d, got %d", id, retrievedRide.ID)
	}
}

func TestUpdateAndDeleteBookRide(t *testing.T) {
	ctx := context.Background()
	repo, err := NewBookingRepository(ctx, "postgresql://postgres:postgres@localhost:5432/luxsuv?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	ride := &data.BookRide{
		YourName:           "Jane Doe",
		Email:              "jane@example.com",
		PhoneNumber:        "098-765-4321",
		RideType:           "per_ride",
		PickupLocation:     "456 Oak St",
		DropoffLocation:    "789 Pine St",
		Date:               "2025-06-24",
		Time:               "15:00",
		NumberOfPassengers: 3,
		NumberOfLuggage:    2,
		AdditionalNotes:    "Please confirm",
	}
	id, err := repo.CreateBookRide(ctx, ride)
	if err != nil {
		t.Fatalf("Failed to create ride: %v", err)
	}
	fmt.Printf("Created ride with ID: %d\n", id)

	updatedRide := &data.BookRide{
		ID:                 id,
		YourName:           "Jane Smith",
		Email:              "jane.smith@example.com",
		PhoneNumber:        "111-222-3333",
		RideType:           "per_ride",
		PickupLocation:     "789 Pine St",
		DropoffLocation:    "123 Maple St",
		Date:               "2025-06-25",
		Time:               "16:00",
		NumberOfPassengers: 4,
		NumberOfLuggage:    3,
		AdditionalNotes:    "Updated note",
	}
	if err := repo.UpdateBookRide(ctx, updatedRide); err != nil {
		t.Fatalf("Failed to update ride: %v", err)
	}
	fmt.Printf("Updated ride with ID: %d\n", id)

	retrievedRide, err := repo.GetBookRideByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get updated ride: %v", err)
	}
	if retrievedRide.YourName != "Jane Smith" {
		t.Errorf("Expected YourName Jane Smith, got %s", retrievedRide.YourName)
	}

	fmt.Printf("Deleting ride with ID: %d\n", id)
	if err := repo.Ping(ctx); err != nil {
		t.Fatalf("Database connection failed: %v", err)
	}
	if err := repo.DeleteBookRide(ctx, id); err != nil {
		t.Fatalf("Failed to delete ride: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("Verifying deletion for ID: %d\n", id)
	newCtx := context.Background()
	_, err = repo.GetBookRideByID(newCtx, id)
	if err == nil {
		t.Errorf("Expected ride to be deleted, but it was found")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Unexpected error during deletion verification: %v", err)
	}
}
