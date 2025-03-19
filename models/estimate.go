package models

import "time"

type Estimate struct {
    ID             int     `json:"id" db:"id"`
    Material       string  `json:"material" db:"material"`
    Length         float64 `json:"length" db:"length"`
    Width          float64 `json:"width" db:"width"`
    Thickness      float64 `json:"thickness" db:"thickness"`
    EdgeFinish     string  `json:"edgeFinish" db:"edge_finish"`
    MaterialCost   float64 `json:"materialCost" db:"material_cost"`
    EdgeFinishCost float64 `json:"edgeFinishCost" db:"edge_finish_cost"`
    LaborCost      float64 `json:"laborCost" db:"labor_cost"`
    TaxRate        float64 `json:"taxRate" db:"tax_rate"`
    Discount       float64 `json:"discount" db:"discount"`
    Cost           float64 `json:"cost" db:"cost"`
    Status         string  `json:"status" db:"status"`
}

type Task struct {
    ID          int       `json:"id" db:"id"`
    EstimateID  int       `json:"estimateId" db:"estimate_id"`
    AssignedTo  string    `json:"assignedTo" db:"assigned_to"`
    DueDate     time.Time `json:"dueDate" db:"due_date"`
    Completed   bool      `json:"completed" db:"completed"`
}

type CustomerInteraction struct {
    ID              int       `json:"id" db:"id"`
    EstimateID      int       `json:"estimateId" db:"estimate_id"`
    InteractionType string    `json:"interactionType" db:"interaction_type"`
    InteractionTime time.Time `json:"interactionTime" db:"interaction_time"`
}