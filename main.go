package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sse-demo/db"
	"sse-demo/models"
	"sync"
	"time"
)

func main() {
	http.HandleFunc("/generate-assets", GenerateAssetsHandler)
	http.HandleFunc("/progress-assets", progressionSSEHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

var (
	activeClients = make(map[string][]*models.Client)
	progressState = make(map[string]int)
	clientsMutex  = sync.Mutex{}
	progressMutex = sync.Mutex{}
)

func GenerateAssetsHandler(w http.ResponseWriter, r *http.Request) {
	projectDB := db.GetProjectFromDB()

	json.NewEncoder(w).Encode(map[string]string{"projectID": projectDB.ID})

	progressMutex.Lock()
	progressState[projectDB.ID] = 0
	progressMutex.Unlock()

	var wg sync.WaitGroup

	for _, asset := range projectDB.Assets {
		wg.Add(1)
		go func(asset models.Asset) {
			defer wg.Done()
			generateAsset(asset, projectDB.ID)
		}(asset)
	}

	go func() {
		wg.Wait()
		progressMutex.Lock()
		progressState[projectDB.ID] = 100
		progressMutex.Unlock()
		broadcast(projectDB.ID, "100")
	}()
}

func generateAsset(asset models.Asset, projectID string) {
	progressIncrement := 0
	switch asset.FileType {
	case models.AudioDescriptionNormal:
		progressIncrement = 10
		time.Sleep(1000 * time.Millisecond)
	case models.TranscriptSRT:
		progressIncrement = 10
		time.Sleep(2000 * time.Millisecond)
	case models.SubtitlesVTT:
		progressIncrement = 10
		time.Sleep(3000 * time.Millisecond)
	case models.ThumbnailJPG:
		progressIncrement = 25
		time.Sleep(8000 * time.Millisecond)
	case models.SignLanguagePictureInPicture:
		progressIncrement = 35
		time.Sleep(15000 * time.Millisecond)
	default:
		log.Printf("Unknown asset type for projectID: %s", projectID)
		return
	}

	progressMutex.Lock()
	progressState[projectID] += progressIncrement
	currentProgress := progressState[projectID]
	progressMutex.Unlock()

	broadcast(projectID, fmt.Sprintf("%d", currentProgress))
}

func broadcast(projectID string, message string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range activeClients[projectID] {
		select {
		case client.Channel <- message:
			log.Printf("Message sent to client for projectID: %s for progress (message): %s", projectID, message)
		default:
			log.Printf("Message dropped for client on projectID: %s for progress (message): %s", projectID, message)
		}
	}
}

func progressionSSEHandler(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("projectID")
	if projectID == "" {
		http.Error(w, "Missing projectID", http.StatusBadRequest)
		return
	}

	// SSE setup
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client := &models.Client{
		ProjectID:   projectID,
		Channel:     make(chan string),
		ConnectTime: time.Now(),
	}

	clientsMutex.Lock()
	activeClients[projectID] = append(activeClients[projectID], client)
	clientsMutex.Unlock()

	go func() {
		<-r.Context().Done()
		unregisterClient(projectID, client)
	}()

	// Stream messages to the client
	for message := range client.Channel {
		jsonMessage, _ := json.Marshal(map[string]string{"progress": message})
		fmt.Fprintf(w, "data: %s\n\n", jsonMessage)
		w.(http.Flusher).Flush()
	}
}

func unregisterClient(projectID string, client *models.Client) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	clients := activeClients[projectID]
	for i, c := range clients {
		if c == client {
			activeClients[projectID] = append(clients[:i], clients[i+1:]...)
			close(client.Channel)
			log.Printf("Client disconnected - projectID: %s, Connection Duration: %v", projectID, time.Since(client.ConnectTime))
			break
		}
	}

	if len(activeClients[projectID]) == 0 {
		delete(activeClients, projectID)
	}
}
