package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type WebhookPayload struct {
	TaskID string `json:"task_id"`
	S3URL  string `json:"s3_url"`
}

var sseClients = make(map[chan WebhookPayload]bool)

func main() {
	http.Handle("/", http.FileServer(http.Dir("static"))) // Assuming your HTML file is in a folder named static
	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/sse", sseHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	for clientMessageChannel := range sseClients {
		clientMessageChannel <- payload
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	messageChannel := make(chan WebhookPayload)
	sseClients[messageChannel] = true

	// Cleanup when the connection is closed
	defer func() {
		delete(sseClients, messageChannel)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)

	for {
		message, open := <-messageChannel
		if !open {
			break
		}

		fmt.Fprintf(w, "data: %s\n\n", json.NewEncoder(w).Encode(message))
		flusher.Flush()
	}
}
