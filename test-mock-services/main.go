package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

type Provider struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Location string    `json:"location"`
	Status   string    `json:"status"`
	GPUs     []GPU     `json:"gpus"`
}

type GPU struct {
	ModelName string `json:"model_name"`
	VRAM      uint64 `json:"vram_mb"`
	IsHealthy bool   `json:"is_healthy"`
}

type Job struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Cost   string `json:"cost"`
}

var providers = make(map[string]*Provider)
var jobs = make(map[string]*Job)

func main() {
	fmt.Println("Starting Test Services...")

	// API Gateway Mock (Port 8090)
	go startAPIGateway()

	// Provider Registry Mock (Port 8091)
	go startProviderRegistry()

	// Billing Service Mock (Port 8092)
	go startBillingService()

	// Storage Service Mock (Port 8093)
	go startStorageService()

	// Wait a moment for services to start
	time.Sleep(2 * time.Second)

	fmt.Println("All test services started!")
	fmt.Printf("API Gateway: http://localhost:%d\n", 8090)
	fmt.Printf("Provider Registry: http://localhost:%d\n", 8091)
	fmt.Printf("Billing Service: http://localhost:%d\n", 8092)
	fmt.Printf("Storage Service: http://localhost:%d\n", 8093)
	fmt.Println("Press Ctrl+C to stop all services")

	// Keep main goroutine alive
	select {}
}

func startAPIGateway() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Authentication
	router.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"token":      "test_jwt_token_12345",
			"user_id":    "user123",
			"username":   "testuser",
			"expires_at": time.Now().Add(24 * time.Hour),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Job submission
	router.HandleFunc("/api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			jobID := uuid.New().String()
			job := &Job{
				ID:     jobID,
				Status: "submitted",
				Cost:   "0.5",
			}
			jobs[jobID] = job

			response := map[string]interface{}{
				"job_id":         jobID,
				"status":         "submitted",
				"estimated_cost": "0.5",
				"message":        "Job submitted successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	// Job status
	router.HandleFunc("/api/v1/jobs/", func(w http.ResponseWriter, r *http.Request) {
		jobID := r.URL.Path[len("/api/v1/jobs/"):]
		if job, exists := jobs[jobID]; exists {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(job)
		} else {
			http.NotFound(w, r)
		}
	})

	fmt.Println("API Gateway starting on :8090")
	log.Fatal(http.ListenAndServe(":8090", router))
}

func startProviderRegistry() {
	mux := http.NewServeMux()

	// Provider registration
	mux.HandleFunc("/api/v1/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var reqData map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqData)

			provider := &Provider{
				ID:       uuid.New(),
				Name:     fmt.Sprintf("%v", reqData["name"]),
				Location: fmt.Sprintf("%v", reqData["location"]),
				Status:   "online",
				GPUs: []GPU{
					{
						ModelName: "Apple M1 Pro",
						VRAM:      16384,
						IsHealthy: true,
					},
				},
			}
			providers[provider.ID.String()] = provider

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(provider)
		} else if r.Method == "GET" {
			var providerList []*Provider
			for _, p := range providers {
				providerList = append(providerList, p)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(providerList)
		}
	})

	// Provider heartbeat
	mux.HandleFunc("/api/v1/providers/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	fmt.Println("Provider Registry starting on :8091")
	log.Fatal(http.ListenAndServe(":8091", mux))
}

func startBillingService() {
	mux := http.NewServeMux()

	// Wallet balance
	mux.HandleFunc("/api/v1/wallet/balance", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"balance":           "10.5",
			"available_balance": "10.5",
			"locked_balance":    "0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Session start
	mux.HandleFunc("/api/v1/billing/session/start", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"session_id":     uuid.New().String(),
			"status":         "started",
			"hourly_rate":    "0.5",
			"estimated_cost": "0.5",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Billing Service starting on :8092")
	log.Fatal(http.ListenAndServe(":8092", mux))
}

func startStorageService() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Storage Service OK")
	})

	mux.HandleFunc("/api/v1/storage/upload", func(w http.ResponseWriter, r *http.Request) {
		// ... existing code ...
	})

	fmt.Println("Storage Service starting on :8093")
	log.Fatal(http.ListenAndServe(":8093", mux))
}
