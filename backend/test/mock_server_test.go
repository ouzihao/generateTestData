package test

import (
	"encoding/json"
	"fmt"
	"generateTestData/backend/config"
	"generateTestData/backend/models"
	"generateTestData/backend/services"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPushToMockServer(t *testing.T) {
	// 1. Setup Mock Server
	var receivedData []map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// The service sends data wrapped in a payload object
		var payload struct {
			Type string                   `json:"type"`
			Data []map[string]interface{} `json:"data"`
		}

		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("Failed to parse request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		receivedData = payload.Data

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	// 2. Setup DB and Service
	dbPath := "test_mock_server.db"
	os.Remove(dbPath)

	config.AppConfig = &config.Config{
		DBPath:      dbPath,
		GenerateDir: ".",
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	models.DB = db

	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 3. Create Task
	mockConfig := map[string]string{
		"url":   ts.URL,
		"token": "test-token",
	}
	mockConfigJSON, _ := json.Marshal(mockConfig)

	jsonSchema := `{"name": "test"}`

	task := models.Task{
		Name:          "Mock Push Test",
		Type:          models.TaskTypeJSON,
		Count:         5,
		JSONSchema:    jsonSchema,
		FieldRules:    "{}", // No special rules, just schema
		OutputType:    models.OutputTypeMockServer,
		Configuration: string(mockConfigJSON),
		OutputPath:    ts.URL, // fallback if config empty, but we set config
	}

	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// 4. Execute Task
	taskService := services.NewTaskService()
	err = taskService.ExecuteTask(task.ID)
	if err != nil {
		t.Fatalf("ExecuteTask failed: %v", err)
	}

	// 5. Wait for completion
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
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

	// 6. Verify received data
	if len(receivedData) != 5 {
		t.Errorf("Expected 5 records, got %d", len(receivedData))
	}

	// Cleanup
	os.Remove(dbPath)
	fmt.Println("TestPushToMockServer Passed!")
}
