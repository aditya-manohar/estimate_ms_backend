// main.go
package main

import (
    "log"
    "net/http"
    "estimate-management-system/db"
    "estimate-management-system/handlers"
    "github.com/gorilla/mux"
    "github.com/rs/cors"
)

func main() {
    // Initialize the database
    if err := db.InitDB(); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Set up the router
    r := mux.NewRouter()

    // Define API routes
    r.HandleFunc("/estimates", handlers.CreateEstimate).Methods("POST")
    r.HandleFunc("/estimates", handlers.GetEstimates).Methods("GET")
    r.HandleFunc("/estimates/{id}", handlers.UpdateEstimate).Methods("PUT")
    r.HandleFunc("/estimates/{id}", handlers.DeleteEstimate).Methods("DELETE")

    // Define tasks routes
    r.HandleFunc("/tasks", handlers.CreateTask).Methods("POST")
    r.HandleFunc("/tasks", handlers.GetTasks).Methods("GET") // Add this line
    r.HandleFunc("/tasks/{id}/update", handlers.UpdateTask).Methods("PUT")

    r.HandleFunc("/customer-interactions", handlers.GetCustomerInteractions).Methods("GET")

    // Enable CORS
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"}, // Allow frontend origin
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    // Wrap the router with CORS
    handler := c.Handler(r)

    // Start the server
    log.Println("Server running on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}