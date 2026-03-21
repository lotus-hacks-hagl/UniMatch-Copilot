package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	loginData := map[string]string{
		"username": "testadmin@unimatch.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(loginData)

	resp, err := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Login request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Login Response:", string(body))

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)

	if result.Data.Token != "" {
		fmt.Println("Access Token:", result.Data.Token)
		
		// Test cases list
		req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/cases?status=all&page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+result.Data.Token)
		casesResp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("Cases request failed: %v", err)
		}
		defer casesResp.Body.Close()
		casesBody, _ := io.ReadAll(casesResp.Body)
		fmt.Println("Cases Response:", string(casesBody))
	}
}
