package repository

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func SavePDFData(name, cpf, email string) {
	// Create data structure
	data := map[string]string{
		"name":         name,
		"cpf_telefone": cpf,
		"email":        email,
		"created_at":   time.Now().Format(time.RFC3339),
	}

	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("error marshaling JSON data: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://sheetdb.io/api/v1/3oadq0rf6skcj", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		log.Printf("request failed with status code: %d", resp.StatusCode)
	}
}
