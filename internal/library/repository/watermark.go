package repository

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func SavePDFData(name, cpf, email string) {
	data := map[string]string{
		"name":         name,
		"cpf_telefone": cpf,
		"email":        email,
		"created_at":   time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("error marshaling JSON data: %v", err)
	}

	req, err := http.NewRequest("POST", "https://sheetdb.io/api/v1/3oadq0rf6skcj", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("request failed with status code: %d", resp.StatusCode)
	}
}
