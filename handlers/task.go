package handlers

import (
    "encoding/json"
    "log"
    "net/http"
"strconv"

    "github.com/gorilla/mux"
    "estimate-management-system/models"
    "estimate-management-system/db"
)

func CreateTask(w http.ResponseWriter, r *http.Request) {
    var task models.Task
    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        log.Println("Error decoding request body:", err)
        http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Received task data: %+v", task) // Add this line

    // Validate input
    if task.EstimateID <= 0 || task.AssignedTo == "" || task.DueDate.IsZero() {
        http.Error(w, "Invalid input: missing or incorrect values", http.StatusBadRequest)
        return
    }

    query := `
        INSERT INTO tasks (estimate_id, assigned_to, due_date, completed)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

    var id int
    err := db.DB.QueryRow(
        query,
        task.EstimateID, task.AssignedTo, task.DueDate, task.Completed,
    ).Scan(&id)
    if err != nil {
        log.Println("Error inserting task:", err)
        http.Error(w, "Failed to create task: "+err.Error(), http.StatusInternalServerError)
        return
    }

    task.ID = id
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}
// GetTasks fetches all tasks
func GetTasks(w http.ResponseWriter, r *http.Request) {
    var tasks []models.Task
    err := db.DB.Select(&tasks, "SELECT * FROM tasks")
    if err != nil {
        log.Println("Error fetching tasks:", err)
        http.Error(w, "Failed to fetch tasks: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}
// In task.go
func UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    taskID, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    var request struct {
        Completed bool `json:"completed"`
    }
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    query := `
        UPDATE tasks
        SET completed = $1
        WHERE id = $2
    `
    _, err = db.DB.Exec(query, request.Completed, taskID)
    if err != nil {
        log.Println("Error updating task:", err)
        http.Error(w, "Failed to update task", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}

// GetCustomerInteractions fetches all customer interactions
func GetCustomerInteractions(w http.ResponseWriter, r *http.Request) {
    var interactions []models.CustomerInteraction
    err := db.DB.Select(&interactions, "SELECT * FROM customer_interactions")
    if err != nil {
        log.Println("Error fetching interactions:", err)
        http.Error(w, "Failed to fetch interactions: "+err.Error(), http.StatusInternalServerError)
        return
    }
    log.Println("Fetched interactions:", interactions) // Debug log
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(interactions)
}