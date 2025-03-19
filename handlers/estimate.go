package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    "estimate-management-system/models"
    "estimate-management-system/db"
)

// CreateEstimate handles creating a new estimate
func CreateEstimate(w http.ResponseWriter, r *http.Request) {
    var estimate models.Estimate

    if err := json.NewDecoder(r.Body).Decode(&estimate); err != nil {
        log.Println("Error decoding request body:", err)
        http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
        return
    }

    // Validate input
    if estimate.Material == "" || estimate.Length <= 0 || estimate.Width <= 0 || estimate.Thickness <= 0 ||
       estimate.EdgeFinish == "" || estimate.MaterialCost < 0 || estimate.EdgeFinishCost < 0 ||
       estimate.LaborCost < 0 || estimate.TaxRate < 0 || estimate.Discount < 0 || estimate.Cost <= 0 {
        http.Error(w, "Invalid input: missing or incorrect values", http.StatusBadRequest)
        return
    }

    query := `
        INSERT INTO estimates (
            material, length, width, thickness, edge_finish, 
            material_cost, edge_finish_cost, labor_cost, tax_rate, discount, 
            cost, status
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        RETURNING id
    `

    var id int
    err := db.DB.QueryRow(
        query, 
        estimate.Material, estimate.Length, estimate.Width, estimate.Thickness, estimate.EdgeFinish,
        estimate.MaterialCost, estimate.EdgeFinishCost, estimate.LaborCost, estimate.TaxRate, estimate.Discount,
        estimate.Cost, estimate.Status,
    ).Scan(&id)
    if err != nil {
        log.Println("Error inserting estimate:", err)
        http.Error(w, "Failed to create estimate: "+err.Error(), http.StatusInternalServerError)
        return
    }

    estimate.ID = id
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(estimate)

    // Create a follow-up task if status is "Sent"
    if estimate.Status == "Sent" {
        go AutomateFollowUp(id)
    }
}

// GetEstimates fetches all estimates
func GetEstimates(w http.ResponseWriter, r *http.Request) {
    var estimates []models.Estimate
    err := db.DB.Select(&estimates, "SELECT * FROM estimates")
    if err != nil {
        log.Println("Error fetching estimates:", err)
        http.Error(w, "Failed to fetch estimates: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(estimates)
}

// GetEstimateByID fetches a specific estimate by ID
func GetEstimateByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    var estimate models.Estimate
    err = db.DB.Get(&estimate, "SELECT * FROM estimates WHERE id = $1", id)
    if err != nil {
        log.Println("Error fetching estimate by ID:", err)
        http.Error(w, "Estimate not found: "+err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(estimate)
}

// UpdateEstimate updates an existing estimate
func UpdateEstimate(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        log.Println("Invalid ID format:", err)
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    var estimate models.Estimate
    if err := json.NewDecoder(r.Body).Decode(&estimate); err != nil {
        log.Println("Error decoding request body:", err)
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Check if the status is being changed to "Sent"
    var existingEstimate models.Estimate
    err = db.DB.Get(&existingEstimate, "SELECT status FROM estimates WHERE id = $1", id)
    if err != nil {
        log.Println("Error fetching existing estimate:", err)
        http.Error(w, "Estimate not found", http.StatusNotFound)
        return
    }

    // If the status is being changed to "Sent", create a follow-up task
    if estimate.Status == "Sent" && existingEstimate.Status != "Sent" {
        go AutomateFollowUp(id) // Create a follow-up task
    }

    // Update the estimate
    query := `
        UPDATE estimates
        SET 
            material = $1, length = $2, width = $3, thickness = $4, edge_finish = $5, 
            material_cost = $6, edge_finish_cost = $7, labor_cost = $8, tax_rate = $9, discount = $10, 
            cost = $11, status = $12
        WHERE id = $13
    `
    _, err = db.DB.Exec(
        query, 
        estimate.Material, estimate.Length, estimate.Width, estimate.Thickness, estimate.EdgeFinish,
        estimate.MaterialCost, estimate.EdgeFinishCost, estimate.LaborCost, estimate.TaxRate, estimate.Discount,
        estimate.Cost, estimate.Status, id,
    )
    if err != nil {
        log.Println("Error updating estimate:", err)
        http.Error(w, "Failed to update estimate", http.StatusInternalServerError)
        return
    }

    // Fetch the updated estimate to return in the response
    var updatedEstimate models.Estimate
    err = db.DB.Get(&updatedEstimate, "SELECT * FROM estimates WHERE id = $1", id)
    if err != nil {
        log.Println("Error fetching updated estimate:", err)
        http.Error(w, "Failed to fetch updated estimate", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(updatedEstimate)
}

// DeleteEstimate deletes an estimate by ID
func DeleteEstimate(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    log.Printf("Received DELETE request for ID: %d", id)

    // Check if the estimate exists before deleting
    var exists bool
    err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM estimates WHERE id = $1)", id).Scan(&exists)
    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    if !exists {
        log.Printf("Estimate with ID %d not found", id)
        http.Error(w, "Estimate not found", http.StatusNotFound)
        return
    }

    // Perform deletion
    query := "DELETE FROM estimates WHERE id = $1"
    result, err := db.DB.Exec(query, id)
    if err != nil {
        log.Printf("Error deleting estimate: %v", err)
        http.Error(w, "Failed to delete estimate: "+err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        log.Printf("No estimate deleted for ID: %d", id)
        http.Error(w, "Estimate not found", http.StatusNotFound)
        return
    }

    log.Printf("Successfully deleted estimate with ID: %d", id)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Deleted successfully"})
}

// AutomateFollowUp creates a follow-up task when an estimate is sent
func AutomateFollowUp(estimateID int) {
    task := models.Task{
        EstimateID: estimateID,
        AssignedTo: "Unassigned", // Default value
        DueDate:    time.Now().Add(24 * time.Hour), // Due in 24 hours
        Completed:  false,
    }

    // Save the task to the database
    _, err := db.DB.Exec(`
        INSERT INTO tasks (estimate_id, assigned_to, due_date, completed)
        VALUES ($1, $2, $3, $4)
    `, task.EstimateID, task.AssignedTo, task.DueDate, task.Completed)
    if err != nil {
        log.Println("Failed to create follow-up task:", err)
        return
    }

    // Log the customer interaction
    _, err = db.DB.Exec(`
        INSERT INTO customer_interactions (estimate_id, interaction_type, interaction_time)
        VALUES ($1, $2, $3)
    `, estimateID, "Reminder Created", time.Now())
    if err != nil {
        log.Println("Failed to log reminder:", err)
    }
}

// AssignTask assigns a task to a sales representative
func AssignTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    taskID, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    var request struct {
        AssignedTo string `json:"assignedTo"`
    }
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    _, err = db.DB.Exec(`
        UPDATE tasks
        SET assigned_to = $1
        WHERE id = $2
    `, request.AssignedTo, taskID)
    if err != nil {
        log.Println("Error assigning task:", err)
        http.Error(w, "Failed to assign task", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Task assigned successfully"})
}

// UpdateTask updates the status of a task
// UpdateTask updates the status of a task
func UpdateTask(w http.ResponseWriter, r *http.Request) {
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

// CheckTaskReminders checks for overdue tasks and sends reminders
func CheckTaskReminders() {
    for {
        // Find tasks that are overdue and not completed
        var tasks []models.Task
        err := db.DB.Select(&tasks, `
            SELECT * FROM tasks
            WHERE due_date < NOW() AND completed = FALSE
        `)
        if err != nil {
            log.Println("Error fetching overdue tasks:", err)
            continue
        }

        for _, task := range tasks {
            // Send a reminder (e.g., email or notification)
            log.Printf("Sending reminder for task ID: %d (Estimate ID: %d)", task.ID, task.EstimateID)

            // Log the interaction
            _, err = db.DB.Exec(`
                INSERT INTO customer_interactions (estimate_id, interaction_type, interaction_time)
                VALUES ($1, $2, $3)
            `, task.EstimateID, "Reminder Sent", time.Now())
            if err != nil {
                log.Println("Failed to log reminder:", err)
            }
        }

        // Sleep for a while before checking again (e.g., every hour)
        time.Sleep(1 * time.Hour)
    }
}