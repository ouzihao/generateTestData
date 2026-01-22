package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	GenerateServerURL = "http://localhost:8070"
	MockServerURL     = "http://localhost:8089"
)

func main() {
	// 1. Login to Mock Server
	token, err := loginToMockServer()
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}
	fmt.Printf("Login successful\n")

	// 2. Create Task in GenerateTestData
	taskID, err := createTask(token)
	if err != nil {
		fmt.Printf("Create task failed: %v\n", err)
		return
	}
	fmt.Printf("Task created, ID: %d\n", taskID)

	// 3. Execute Task
	err = executeTask(taskID)
	if err != nil {
		fmt.Printf("Execute task failed: %v\n", err)
		return
	}
	fmt.Println("Task execution started...")

	// 4. Wait for task completion (poll status)
	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)
		status, err := getTaskStatus(taskID)
		if err != nil {
			fmt.Printf("Get task status failed: %v\n", err)
			return
		}
		fmt.Printf("Task status: %s\n", status)
		if status == "completed" {
			break
		}
		if status == "failed" {
			fmt.Println("Task failed!")
			return
		}
	}

	// 5. Verify data in Mock Server
	err = verifyMockData(token)
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
		return
	}
	fmt.Println("Integration verification SUCCESS!")
}

func loginToMockServer() (string, error) {
	payload := map[string]string{
		"username": "admin",
		"password": "password123",
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(MockServerURL+"/api/auth/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return "Bearer " + result.AccessToken, nil
}

func createTask(token string) (int, error) {
	// Configuration for pushing to Mock Server
	config := map[string]string{
		"url":   MockServerURL + "/api/import",
		"type":  "users",
		"token": token,
	}
	configBytes, _ := json.Marshal(config)

	// Field rules for users
	fieldRules := map[string]interface{}{
		"username": map[string]interface{}{
			"type": "regex",
			"parameters": map[string]interface{}{
				"pattern": "[a-z]{8}",
			},
		},
		"email": map[string]interface{}{
			"type": "regex",
			"parameters": map[string]interface{}{
				"pattern": "[a-z]{5,10}@[a-z]{5}\\.com",
			},
		},
		"role": map[string]interface{}{
			"type": "fixed",
			"parameters": map[string]interface{}{
				"value": "user",
			},
		},
	}
	rulesJSON, _ := json.Marshal(fieldRules)

	// JSON Schema
	jsonSchema := map[string]string{
		"username": "",
		"email":    "",
		"role":     "",
	}
	schemaBytes, _ := json.Marshal(jsonSchema)

	task := map[string]interface{}{
		"name":          "Integration Test Task",
		"type":          "json",
		"count":         5,
		"outputType":    "mock_server",
		"configuration": string(configBytes),
		"fieldRules":    string(rulesJSON),
		"jsonSchema":    string(schemaBytes),
		"outputPath":    MockServerURL + "/api/import", // Compatibility
	}

	data, _ := json.Marshal(task)
	resp, err := http.Post(GenerateServerURL+"/api/tasks", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Data.ID, nil
}

func executeTask(id int) error {
	resp, err := http.Post(fmt.Sprintf("%s/api/tasks/%d/execute", GenerateServerURL, id), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func getTaskStatus(id int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/tasks/%d", GenerateServerURL, id))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.Status, nil
}

func verifyMockData(token string) error {
	req, _ := http.NewRequest("GET", MockServerURL+"/api/users", nil)
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []map[string]interface{} `json:"items"`
		Total int                      `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	users := result.Items
	fmt.Printf("Found %d users in Mock Server\n", len(users))
	// We expect at least 5 new users plus existing ones.
	// Initial mock server has 4 users (admin, user1, user2, guest).
	// So 4 + 5 = 9.
	if len(users) >= 9 {
		return nil
	}
	return fmt.Errorf("expected at least 9 users, got %d", len(users))
}
