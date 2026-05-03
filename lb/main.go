package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getenvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, _ := strconv.Atoi(v)
	return i
}

func main() {
	flag.Parse()

	// Simple env-driven config
	dt := getenvFloat("LB_ALPHA_DT", 0.1)
	tau := getenvFloat("LB_ALPHA_TAU", 10.0)
	alpha := AlphaFromDtTau(dt, tau)
	wLat := getenvFloat("LB_WEIGHT_LAT", 0.7)
	wConn := getenvFloat("LB_WEIGHT_CONN", 0.3)
	errPenaltyMs := getenvFloat("LB_ERR_PENALTY_MS", 500)

	backends := []string{
		os.Getenv("BACKEND1"),
		os.Getenv("BACKEND2"),
		os.Getenv("BACKEND3"),
	}
	var bks []*Backend
	for _, raw := range backends {
		if raw == "" {
			continue
		}
		b, err := NewBackend(raw)
		if err != nil {
			log.Fatalf("invalid backend %s: %v", raw, err)
		}
		bks = append(bks, b)
	}
	if len(bks) == 0 {
		log.Fatalf("no backends configured")
	}

	lb := NewLoadBalancer(bks, alpha, wLat, wConn, errPenaltyMs)
	hcPath := os.Getenv("LB_HEALTHCHECK_PATH")
	if hcPath == "" {
		hcPath = "/health"
	}
	hcInterval := getenvInt("LB_HEALTHCHECK_INTERVAL_SEC", 5)
	if envInt := os.Getenv("LB_HEALTHCHECK_INTERVAL"); envInt != "" {
		// allow duration like 5s
		d, err := time.ParseDuration(envInt)
		if err == nil {
			hcInterval = int(d.Seconds())
		}
	}
	lb.StartHealthChecks(time.Duration(hcInterval)*time.Second, hcPath)

	http.HandleFunc("/", lb.ProxyHandler)
	http.HandleFunc(hcPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("LB_PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	readTimeout := 5 * time.Second
	writeTimeout := 20 * time.Second
	idleTimeout := 60 * time.Second

	if v := os.Getenv("LB_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			readTimeout = d
		}
	}
	if v := os.Getenv("LB_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			writeTimeout = d
		}
	}
	if v := os.Getenv("LB_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			idleTimeout = d
		}
	}

	srv := &http.Server{Addr: addr, ReadTimeout: readTimeout, WriteTimeout: writeTimeout, IdleTimeout: idleTimeout}
	log.Printf("Predictive Sentinel escuchando en %s (alpha=%.6f)", addr, alpha)
	log.Fatal(srv.ListenAndServe())
}
