package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// AssetProgress keeps track of the progress of each assetID
var assetProgress = make(map[string]int)

// Client holds information about each client connection
type Client struct {
	AssetID     string
	Channel     chan string
	ConnectTime time.Time
}

var activeClients = make(map[string][]*Client)
var clientsMutex = sync.Mutex{}

func main() {
	http.HandleFunc("/generate", GenerateHandler)
	http.HandleFunc("/sse", SSEHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// GenerateHandler generates a new assetID and initiates the generation process
func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	assetID := strconv.Itoa(rand.Intn(100000))
	assetProgress[assetID] = 0

	json.NewEncoder(w).Encode(map[string]string{"assetID": assetID})

	log.Printf("Generated assetID: %s and started tracking progress.", assetID)

	go simulateProgress(assetID)
}

func SSEHandler(w http.ResponseWriter, r *http.Request) {
	assetID := r.URL.Query().Get("assetID")
	if assetID == "" {
		http.Error(w, "Missing assetID", http.StatusBadRequest)
		return
	}

	// Get client IP and User-Agent
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	userAgent := r.Header.Get("User-Agent")
	log.Printf("Client connected - IP: %s, User-Agent: %s, AssetID: %s", ip, userAgent, assetID)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := &Client{
		AssetID:     assetID,
		Channel:     make(chan string),
		ConnectTime: time.Now(),
	}

	clientsMutex.Lock()
	activeClients[assetID] = append(activeClients[assetID], client)
	clientsMutex.Unlock()

	go func() {
		<-r.Context().Done() // This will be triggered when the client disconnects
		unregisterClient(assetID, client)
	}()

	for message := range client.Channel {
		fmt.Fprintf(w, "data: %s\n\n", message)
		w.(http.Flusher).Flush()
	}
}

func simulateProgress(assetID string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		assetProgress[assetID] += 10

		// Log progress at key intervals
		if assetProgress[assetID] == 50 || assetProgress[assetID] == 100 {
			log.Printf("AssetID %s progress reached %d%%", assetID, assetProgress[assetID])
		}

		data, err := json.Marshal(map[string]interface{}{
			"assetID":  assetID,
			"progress": assetProgress[assetID],
		})
		if err != nil {
			log.Println("Error encoding JSON:", err)
			return
		}

		broadcast(assetID, string(data))

		if assetProgress[assetID] >= 100 {
			log.Printf("AssetID %s generation completed.", assetID)
			broadcast(assetID, fmt.Sprintf(`{"assetID": "%s", "progress": "completed"}`, assetID))
			delete(assetProgress, assetID)
			break
		}
	}
}

func broadcast(assetID, message string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range activeClients[assetID] {
		client.Channel <- message
	}
}

func unregisterClient(assetID string, client *Client) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	clients := activeClients[assetID]
	for i, c := range clients {
		if c == client {
			activeClients[assetID] = append(clients[:i], clients[i+1:]...)
			close(client.Channel)
			log.Printf("Client disconnected - AssetID: %s, Connection Duration: %v", assetID, time.Since(client.ConnectTime))
			break
		}
	}

	if len(activeClients[assetID]) == 0 {
		delete(activeClients, assetID)
	}
}
