package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}
	latStr := os.Getenv("SIMULATED_LATENCY_MS")
	lat := 50
	if latStr != "" {
		if v, err := strconv.Atoi(latStr); err == nil {
			lat = v
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate variable processing time
		jitter := rand.Intn(lat/2 + 1)
		time.Sleep(time.Duration(lat+jitter) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"ok\":true, \"lat_ms\":%d}", lat+jitter)
	})

	http.HandleFunc("/heavy", func(w http.ResponseWriter, r *http.Request) {
		// heavier path
		time.Sleep(time.Duration(lat*3) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("heavy done"))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Backend listening %s (simulated latency %dms)", addr, lat)
	log.Fatal(http.ListenAndServe(addr, nil))
}
