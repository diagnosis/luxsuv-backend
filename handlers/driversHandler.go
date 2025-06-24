package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"log"
	"luxsuv-backend/repository"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type DriverAuthMiddleware struct {
	repo   *repository.BookingRepository
	secret []byte
}

func NewDriverAuthMiddleware(repo *repository.BookingRepository) *DriverAuthMiddleware {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		log.Fatalf("JWT_SECRET environment variable not set")
	}
	return &DriverAuthMiddleware{
		repo:   repo,
		secret: secret,
	}
}

func (a *DriverAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return a.secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role != "driver" {
			http.Error(w, "Access denied: driver role required", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", int64(claims["id"].(float64)))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetupDriverRouter(repo *repository.BookingRepository) *chi.Mux {
	r := chi.NewRouter()
	authMiddleware := NewDriverAuthMiddleware(repo)

	// Login endpoint
	r.Post("/login", loginHandler(repo, authMiddleware))

	// Protected driver endpoints
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Middleware)

		r.Get("/book-rides", listAllBookRides(repo))
		r.Get("/book-ride/{id}", getBookRide(repo))
		r.Delete("/book-ride/{id}", deleteBookRide(repo))
	})

	return r
}

func loginHandler(repo *repository.BookingRepository, middleware *DriverAuthMiddleware) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
			return
		}

		user, err := repo.GetUserByCredentials(r.Context(), creds.Username, creds.Password)
		if err != nil {
			if err == pgx.ErrNoRows {
				respondError(w, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
				return
			}
			respondError(w, http.StatusInternalServerError, fmt.Errorf("login failed: %w", err))
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenString, err := token.SignedString(middleware.secret) // Use middleware parameter
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to generate token: %w", err))
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"token": tokenString})
	}
}

func getBookRide(repo *repository.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("invalid booking ID: %w", err))
			return
		}

		ride, err := repo.GetBookRideByID(ctx, id)
		if err != nil {
			if err == pgx.ErrNoRows {
				respondError(w, http.StatusNotFound, fmt.Errorf("ride booking not found: %d", id))
				return
			}
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to get ride booking: %w", err))
			return
		}

		respondJSON(w, http.StatusOK, ride)
	}
}

func deleteBookRide(repo *repository.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Errorf("invalid booking ID: %w", err))
			return
		}

		if err := repo.DeleteBookRide(ctx, id); err != nil {
			if err == pgx.ErrNoRows {
				respondError(w, http.StatusNotFound, fmt.Errorf("ride booking not found: %d", id))
				return
			}
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to delete ride booking: %w", err))
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "Ride booking deleted successfully"})
	}
}

func listAllBookRides(repo *repository.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rides, err := repo.ListBookRides(ctx)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Errorf("failed to list all ride bookings: %w", err))
			return
		}
		respondJSON(w, http.StatusOK, rides)
	}
}
