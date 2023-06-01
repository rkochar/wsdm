package main

import (
	"fmt"
	"net/http"
)

var eventCh chan struct{}

func main() {
	// Initialize the event channel
	eventCh = make(chan struct{})

	// Register the HTTP handler
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/release", handleRequestRelease)

	// Start the HTTP server
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}
	}()

	// Wait for the event to occur
	<-eventCh

	// Respond to the HTTP request
	fmt.Println("Event occurred, sending response...")
	http.Error(nil, "Event occurred", http.StatusOK)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Defer the response until the event occurs
	fmt.Println("Received request, deferring response...")
	go func() {
		<-eventCh
		fmt.Println("Event occurred, sending response...")
		http.Error(w, "Event occurred", http.StatusOK)
	}()
}
func handleRequestRelease(w http.ResponseWriter, r *http.Request) {
	eventCh <- struct{}{}
}
