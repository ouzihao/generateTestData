package test

import (
	"encoding/json"
	"fmt"
	"generateTestData/backend/config"
	"generateTestData/backend/models"
	"generateTestData/backend/services"
	"os"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDBLookup(t *testing.T) {
	// Setup
	dbPath := "test_lookup.db"
	os.Remove(dbPath) // Clean up previous run

	// Initialize Config
	config.AppConfig = &config.Config{
		DBPath:      dbPath,
		GenerateDir: ".",
	}

	// Initialize DB
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	models.DB = db

	// Auto migrate
	err = db.AutoMigrate(&models.DataSource{}, &models.Task{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 1. Create a dummy table for lookup
	err = db.Exec("CREATE TABLE IF NOT EXISTS source_users (id INTEGER PRIMARY KEY, username TEXT)").Error
	if err != nil {
		t.Fatalf("Failed to create source table: %v", err)
	}

	// 2. Insert dummy data
	sourceUsers := []string{"alice", "bob", "charlie"}
	for _, u := range sourceUsers {
		db.Exec("INSERT INTO source_users (username) VALUES (?)", u)
	}

	// 3. Create DataSource record pointing to this same DB (SQLite)
	ds := models.DataSource{
		Name:     "Self SQLite",
		Type:     "sqlite",
		Database: dbPath, // For sqlite, we use Database field as path in openConnection
	}
	if err := db.Create(&ds).Error; err != nil {
		t.Fatalf("Failed to create datasource: %v", err)
	}

	// 4. Create Task with db_lookup rule
	// We want to generate a JSON file with a field "selected_user" picking from source_users.username

	ruleParams := map[string]interface{}{
		"dataSourceId": ds.ID,
		"tableName":    "source_users",
		"columnName":   "username",
	}

	fieldRules := map[string]interface{}{
		"selected_user": map[string]interface{}{
			"type":       "db_lookup",
			"parameters": ruleParams,
		},
	}

	fieldRulesJSON, _ := json.Marshal(fieldRules)

	jsonSchema := `{"selected_user": "placeholder"}`

	outputFile := "test_lookup_output.json"
	task := models.Task{
		Name:         "Lookup Test",
		Type:         models.TaskTypeJSON,
		Count:        10,
		JSONSchema:   jsonSchema,
		FieldRules:   string(fieldRulesJSON),
		OutputType:   models.OutputTypeJSON,
		OutputPath:   outputFile,
		DataSourceID: &ds.ID,
	}

	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// 5. Execute Task
	taskService := services.NewTaskService()

	err = taskService.ExecuteTask(task.ID)
	if err != nil {
		t.Fatalf("ExecuteTask failed: %v", err)
	}

	// Wait for task completion
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	completed := false
	for {
		select {
		case <-timeout:
			t.Fatalf("Task timeout")
		case <-ticker.C:
			var currentTask models.Task
			if err := db.First(&currentTask, task.ID).Error; err != nil {
				continue
			}
			if currentTask.Status == models.TaskStatusCompleted {
				completed = true
				break
			}
			if currentTask.Status == models.TaskStatusFailed {
				t.Fatalf("Task failed: %s", currentTask.ErrorMsg)
			}
		}
		if completed {
			break
		}
	}

	// 6. Verify Output
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(content, &results); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}

	for i, r := range results {
		val, ok := r["selected_user"].(string)
		if !ok {
			t.Errorf("Row %d: selected_user is not string", i)
			continue
		}

		found := false
		for _, u := range sourceUsers {
			if val == u {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Row %d: value '%s' not found in source users %v", i, val, sourceUsers)
		}
	}

	// Cleanup
	os.Remove(outputFile)
	os.Remove(dbPath)
	fmt.Println("TestDBLookup Passed!")
}
