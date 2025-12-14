package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"arturgudiev/dashboard/ent"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string
	// Default PostgreSQL connection: postgresql://postgres:postgres@localhost/dashboard
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/dashboard?sslmode=disable"
	}

	// Create Ent client
	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	// Run the auto migration tool
	// If tables already exist, this will handle it gracefully
	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		// Check if the error is because table already exists, permission issue, or schema mismatch
		// If table exists, we can continue - Ent will work with existing tables
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "already exists") || 
		   strings.Contains(errMsg, "permission denied") ||
		   strings.Contains(errMsg, "unexpected attribute change") ||
		   strings.Contains(errMsg, "expect identity") {
			log.Printf("Warning: Schema migration had issues: %v", err)
			
			log.Println("Continuing with existing schema...")
		} else {
			log.Fatalf("failed creating schema resources: %v", err)
		}
	} else {
		log.Println("Schema migration completed successfully")
	}

	// Setup routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Dashboard server"})
	})

	http.HandleFunc("/tests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Query all tests
		tests, err := client.Test.Query().All(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert to JSON response
		type TestResponse struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Tags []string `json:"tags"`
		}

		response := make([]TestResponse, len(tests))
		for i, t := range tests {
			response[i] = TestResponse{
				ID:   t.ID,
				Name: t.Name,
				Tags: t.Tags,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

