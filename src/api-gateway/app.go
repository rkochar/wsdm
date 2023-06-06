package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Create a map of channels
var channelMap = make(map[string]chan int)

// Create a mutex for synchronization
var mutex sync.Mutex

const ORDER_SERVICE = "http://order-service:5000/checkout/"
const TIMEOUT_SECONDS = 100

func main() {
	defer cleanup()

	router := mux.NewRouter()
	router.HandleFunc("/{order_id}", checkoutHandler)
	router.HandleFunc("/release/{order_id}/{status}", unblockCheckout)
	router.HandleFunc("/", homeHandler)

	port := os.Getenv("PORT")
	fmt.Printf("\nCurrent port is: %s", port)
	if port == "" {
		port = "5000"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("\nStarting api gateway service at %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the home page..!")
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	channelCreationStatus := createChannel(orderID)
	if channelCreationStatus {
		log.Printf("\nNew Order call %s received", orderID)
		immediateResponseCode := routeCheckoutCall(orderID)
		if immediateResponseCode != http.StatusOK { // saga start went fine
			w.WriteHeader(http.StatusBadRequest)
		} else {
			select {
			case finalResponseCode := <-channelMap[orderID]:
				w.WriteHeader(finalResponseCode)
				log.Printf("\nReceived status %d for order %s:", finalResponseCode, orderID)
			case <-time.After(TIMEOUT_SECONDS * time.Second):
				log.Printf("\nTimeout occurred for order connection", orderID)
			}

		}
		deleteChannel(orderID)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func unblockCheckout(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	orderID := vars["order_id"]
	status, err := strconv.Atoi(vars["status"])
	log.Printf("\nTrying to unlock order: %s with status %d", orderID, status)
	if err != nil {
		log.Println("Failed to convert string to integer:", err, orderID, vars["status"])
		releaseChannel(orderID, http.StatusBadRequest)
		return
	}
	releaseChannel(orderID, status)
	w.WriteHeader(http.StatusOK)
	return
}

func createChannel(orderId string) bool {
	mutex.Lock()
	ch := channelMap[orderId]
	status := true
	if ch == nil {
		channelMap[orderId] = make(chan int)
	} else {
		// Order saga already in progress, cant place again
		log.Print("Saga already in progress dont start again", orderId)
		status = false
	}
	mutex.Unlock()
	return status
}

func deleteChannel(orderId string) {
	mutex.Lock()
	delete(channelMap, orderId)
	mutex.Unlock()

}

func releaseChannel(orderId string, status int) {
	mutex.Lock()
	ch := channelMap[orderId]
	if ch == nil {
		log.Print("Connection does not exist for order id ", orderId)
		mutex.Unlock()
		return
	} else {
		ch <- status
		mutex.Unlock()
		log.Printf("\nUnlocked order: %s with status %d", orderId, status)
	}

}

func routeCheckoutCall(orderID string) int {
	backendURL := ORDER_SERVICE + orderID
	resp, err := http.Get(backendURL)
	if err != nil {
		log.Printf("\nFailed to make service call: %v", err)
		return http.StatusBadRequest
	}
	return resp.StatusCode

}

func cleanup() {
	// Cleanup and close channels
	mutex.Lock()
	for _, ch := range channelMap {
		close(ch)
	}
	channelMap = nil
	mutex.Unlock()
}
