package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://backend-prod:80"
	}

	intervalStr := os.Getenv("INTERVAL")
	if intervalStr == "" {
		intervalStr = "5s"
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Fatalf("Invalid INTERVAL format: %v", err)
	}

	endpoints := []string{
		"/config",
		"/pods",
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	log.Printf("Traffic generator started")
	log.Printf("Backend URL: %s", backendURL)
	log.Printf("Interval: %s", interval)
	log.Printf("Endpoints: %v", endpoints)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial request
	makeRequests(client, backendURL, endpoints)

	for range ticker.C {
		makeRequests(client, backendURL, endpoints)
	}
}

func makeRequests(client *http.Client, baseURL string, endpoints []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, endpoint := range endpoints {
		url := baseURL + endpoint

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			log.Printf("Error creating request for %s: %v", url, err)
			continue
		}

		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)

		if err != nil {
			log.Printf("Error calling %s: %v (duration: %v)", url, err, duration)
			continue
		}
		defer resp.Body.Close()

		log.Printf("GET %s -> %d (duration: %v)", url, resp.StatusCode, duration)
	}
}
