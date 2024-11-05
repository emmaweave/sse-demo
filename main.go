package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

// AssetProgress keeps track of the progress of each assetID
var assetProgress = make(map[string]int)

// StartHandler generates a new assetID and initiates the generation process
func StartHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a random assetID
	assetID := strconv.Itoa(rand.Intn(100000))
	assetProgress[assetID] = 0 // Initialize progress for this assetID

	// Return the generated assetID to the frontend
	json.NewEncoder(w).Encode(map[string]string{"assetID": assetID})

	fmt.Printf("Started generation for assetID: %s\n", assetID)
}

// SSEHandler handles the Server-Sent Events connection
func SSEHandler(w http.ResponseWriter, r *http.Request) {
	assetID := r.URL.Query().Get("assetID")
	if assetID == "" {
		http.Error(w, "Missing assetID", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if _, exists := assetProgress[assetID]; exists {
			assetProgress[assetID] += 10

			data, err := json.Marshal(map[string]interface{}{
				"assetID":  assetID,
				"progress": assetProgress[assetID],
			})
			if err != nil {
				log.Println("Error encoding JSON:", err)
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()

			if assetProgress[assetID] >= 100 {
				fmt.Fprintf(w, "data: %s\n\n", `{"assetID": "`+assetID+`", "progress": "completed"}`)
				w.(http.Flusher).Flush()
				break
			}
		} else {
			break
		}
	}
}

func main() {
	http.HandleFunc("/generate", StartHandler) // Endpoint to start generation
	http.HandleFunc("/sse", SSEHandler)        // SSE endpoint

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for local testing
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
