package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"govmmonitoring/monitor"
)

var (
	currentStats *monitor.SystemStats
	statsMutex   sync.RWMutex
)

func main() {
	// Initialize stats immediately
	updateStats()

	// Start background ticker to collect stats every 3 seconds
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for range ticker.C {
			updateStats()
		}
	}()

	// Serve static files from the "assets" directory
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/", fs)

	// API endpoint for stats
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api.json", statsHandler)

	port := ":8993"
	log.Printf("Starting server on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func updateStats() {
	stats, err := monitor.GetStats()
	if err != nil {
		log.Printf("Error collecting stats: %v", err)
		return
	}

	statsMutex.Lock()
	currentStats = stats
	statsMutex.Unlock()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	statsMutex.RLock()
	stats := currentStats
	statsMutex.RUnlock()

	if stats == nil {
		http.Error(w, "Stats not available yet", http.StatusServiceUnavailable)
		return
	}

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Error encoding stats", http.StatusInternalServerError)
	}
}
