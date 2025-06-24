LuxSUV Backend
Welcome to the LuxSUV Backend, a robust Go-based RESTful API designed to manage driver and rider services, including user authentication, ride bookings, and ride management. This application leverages the go-chi framework for routing, jackc/pgx for PostgreSQL interactions, and goose for database migrations, with a Neon-hosted PostgreSQL database. It is currently deployed on Fly.io.
Overview

Language: Go (version 1.24.4 or later recommended)
Framework: go-chi/chi (v5.2.2)
Database: PostgreSQL (hosted on Neon)
Authentication: JWT-based with golang-jwt/jwt (v5.2.2)
Dependencies: github.com/jackc/pgx/v5 (v5.7.5), github.com/joho/godotenv (v1.5.1), and others
Deployment: Hosted at https://luxsuv-backend.fly.dev

Features

Driver authentication with JWT tokens.
Rider ride booking and management.
CRUD operations for ride bookings (Create, Read, Update, Delete).
Role-based access control (driver-only endpoints).

Prerequisites

Go: Version 1.24 or higher (go install).
cURL: For API testing.
Goose: For database migrations (go install github.com/pressly/goose/v3/cmd/goose@latest).
Fly CLI: For deployment and secret management (curl -L https://fly.io/install.sh | sh).
Git: For version control.
PostgreSQL Client: Optional, for local testing (e.g., psql).
Local PostgreSQL: For local development (install via Homebrew: brew install postgresql or equivalent).

Setup
Local Development

Clone the Repository:git clone https://github.com/yourusername/luxsuv-backend.git
cd luxsuv-backend


Install Dependencies:go mod tidy


Install Local PostgreSQL:
Install PostgreSQL (e.g., via Homebrew on macOS):brew install postgresql
brew services start postgresql


Create a local database:createdb luxsuv_local




Configure Environment:
Create a .env file in the root directory with your local setup (example):DATABASE_URL=postgresql://localhost/luxsuv_local?sslmode=disable
JWT_SECRET=<your-secure-jwt-secret>


Replace <your-secure-jwt-secret> with a random string (e.g., generated via openssl rand -base64 32).
Keep .env out of version control (add to .gitignore).


Load environment variables:source .env




Apply Migrations:
Ensure the db/migrations directory contains your SQL migration files.
Run migrations against the local database:goose -dir db/migrations postgres "$DATABASE_URL" up




Build and Run Locally:go build -o luxsuv-backend
./luxsuv-backend


Expected: "Server started on :8080".
Access at http://localhost:8080.



Deployment with Fly.io

Initialize Fly.io:fly launch


App name: luxsuv-backend
Region: us-west-2
Answer No to creating a PostgreSQL database.


Set Secrets:
Use Fly CLI to set your database URL and JWT secret:fly secrets set DATABASE_URL=<your-neon-database-url>
fly secrets set JWT_SECRET=<your-jwt-secret>


Note: Replace <your-neon-database-url> and <your-jwt-secret> with secure values from your Neon dashboard and a generated secret.


Deploy:GOOS=linux GOARCH=amd64 go build -o luxsuv-backend ./cmd
fly deploy


Expected: URL like https://luxsuv-backend.fly.dev.



Usage
The API is accessible at https://luxsuv-backend.fly.dev (deployed) or http://localhost:8080 (local). Below are detailed examples for each endpoint.
Driver Endpoints

Login:

Method: POST
Endpoint: /driver/login
Request:curl -X POST https://luxsuv-backend.fly.dev/driver/login \
-H "Content-Type: application/json" \
-d '{"username":"eroldriver","password":"sazanavi123"}'


Response: {"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."} (status 200)
Notes: Returns a JWT token valid for 24 hours.


List All Booked Rides (Protected):

Method: GET
Endpoint: /book-rides
Request:curl -H "Authorization: Bearer <token>" https://luxsuv-backend.fly.dev/book-rides


Response: [{"id":1,"your_name":"John Doe",...}] (status 200)
Notes: Requires a valid driver JWT token in the Authorization header.


Get Booked Ride by ID (Protected):

Method: GET
Endpoint: /book-ride/{id}
Request:curl -H "Authorization: Bearer <token>" https://luxsuv-backend.fly.dev/book-ride/1


Response: {"id":1,"your_name":"John Doe",...} (status 200) or 404 if not found.


Delete Booked Ride by ID (Protected):

Method: DELETE
Endpoint: /book-ride/{id}
Request:curl -X DELETE -H "Authorization: Bearer <token>" https://luxsuv-backend.fly.dev/book-ride/1


Response: {"message":"Ride booking deleted successfully"} (status 200) or 404 if not found.



Rider Endpoints

Root Check:

Method: GET
Endpoint: /
Request:curl https://luxsuv-backend.fly.dev/


Response: {"message":"Hello world"} (status 200).


Book a Ride:

Method: POST
Endpoint: /book-ride
Request:curl -X POST https://luxsuv-backend.fly.dev/book-ride \
-H "Content-Type: application/json" \
-d '{
"your_name": "John Doe",
"email": "john@example.com",
"phone_number": "123-456-7890",
"ride_type": "hourly",
"pickup_location": "123 Main St",
"dropoff_location": "456 Elm St",
"date": "2025-06-24",
"time": "09:00",
"number_of_passengers": 2,
"number_of_luggage": 1,
"additional_notes": "Please arrive early"
}'


Response: {"message":"Ride booking created successfully","id":1} (status 201)
Notes: Validates ride_type as hourly or per_ride.


Update Booked Ride by ID:

Method: PUT
Endpoint: /book-ride/{id}
Request:curl -X PUT https://luxsuv-backend.fly.dev/book-ride/1 \
-H "Content-Type: application/json" \
-d '{
"your_name": "Jane Doe",
"email": "jane@example.com",
"phone_number": "098-765-4321",
"ride_type": "per_ride",
"pickup_location": "789 Oak St",
"dropoff_location": "321 Pine St",
"date": "2025-06-25",
"time": "10:00",
"number_of_passengers": 1,
"number_of_luggage": 0,
"additional_notes": "No luggage"
}'


Response: {"message":"Ride booking updated successfully"} (status 200) or 404 if not found.


List Booked Rides by Email:

Method: GET
Endpoint: /book-rides
Request:curl "https://luxsuv-backend.fly.dev/book-rides?email=john@example.com"


Response: [{"id":1,"your_name":"John Doe",...}] (status 200) or 400 if email is missing.



Deployment

Fly.io Configuration: Uses fly.toml with heroku/builder:24 buildpack.
Build Command: GOOS=linux GOARCH=amd64 go build -o luxsuv-backend ./cmd
Deploy Command: fly deploy

Development

Project Structure:
cmd/: Contains main.go for the application entry point.
handlers/: Defines API routes and handlers (e.g., driverHandler.go, riderHandler.go).
repository/: Manages database interactions (e.g., bookingRepository.go).
data/: Contains data models (e.g., bookRide.go, user.go).
db/migrations/: Stores SQL migration files.


Running Tests: Add tests in a tests/ directory and run with go test ./....
Contributing: Fork the repository, create a feature branch, and submit a pull request. Add developers via the Fly.io dashboard if needed at https://fly.io/dashboard/safa-demirkan/apps/luxsuv-backend/team.
Adding Developers: Invite collaborators through the Fly.io dashboard under the app’s team settings.

Troubleshooting

Database Connection: Ensure DATABASE_URL is set and migrations are applied. Check fly logs for errors.
Fly CLI: If fly is not found, run source ~/.zshrc after adding export PATH="/Users/safademirkan/.fly/bin:$PATH" to .zshrc.
Build Failures: Verify Go version and dependencies with go mod tidy.
API Errors: Use fly logs to debug endpoint issues. Ensure JWT tokens are valid for protected routes.

