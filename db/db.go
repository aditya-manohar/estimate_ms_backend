package db

import (
    "fmt"
    _ "github.com/lib/pq"
    "github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitDB() error {
    connStr := "user=postgres dbname=estimate_system sslmode=disable password=getout612"
    var err error
    DB, err = sqlx.Connect("postgres", connStr)
    if err != nil {
        return fmt.Errorf("failed to connect to database: %v", err)
    }

    // Create estimates table if it doesn't exist
    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS estimates (
            id SERIAL PRIMARY KEY,
            material TEXT,
            length FLOAT,
            width FLOAT,
            thickness FLOAT,
            edge_finish TEXT,
            material_cost FLOAT,
            edge_finish_cost FLOAT,
            labor_cost FLOAT,
            tax_rate FLOAT,
            discount FLOAT,
            cost FLOAT,
            status TEXT
        );
    `)
    if err != nil {
        return fmt.Errorf("failed to create estimates table: %v", err)
    }

    // Create tasks table if it doesn't exist
    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS tasks (
            id SERIAL PRIMARY KEY,
            estimate_id INTEGER REFERENCES estimates(id) ON DELETE CASCADE,
            assigned_to TEXT,
            due_date TIMESTAMP,
            completed BOOLEAN DEFAULT FALSE
        );
    `)
    if err != nil {
        return fmt.Errorf("failed to create tasks table: %v", err)
    }

    // Create customer_interactions table if it doesn't exist
    _, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS customer_interactions (
            id SERIAL PRIMARY KEY,
            estimate_id INTEGER REFERENCES estimates(id) ON DELETE CASCADE,
            interaction_type TEXT,
            interaction_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
    if err != nil {
        return fmt.Errorf("failed to create customer_interactions table: %v", err)
    }

    return nil
}