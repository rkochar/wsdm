package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var cp sync.Mutex
var channelMap = make(map[string](chan int))

const ORDER_SERVICE = "http://order-service:5000/checkout/"

func main() {

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
	log.Print("Checkout handler called")

	order_id := mux.Vars(r)["order_id"]
	immediateResp := routeCheckoutCall(order_id)
	if immediateResp == http.StatusBadRequest {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		created_channel := createChannel(order_id)
		if created_channel {
			select {
			case status := <-channelMap[order_id]:
				log.Printf("Channel released for order: %s and status $d", order_id, status)
				w.WriteHeader(status)
			}
		} else {
			log.Printf("Channel not created for order: %s", order_id)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

}

func unblockCheckout(w http.ResponseWriter, r *http.Request) {
	log.Print("Unblock checkout handler called")
	order_id := mux.Vars(r)["order_id"]
	status := mux.Vars(r)["status"]
	statusi, err := strconv.Atoi(status)
	if err != nil {
		log.Printf("Error converting status to int: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s := releaseChannel(order_id, statusi)
	if s {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

}

func createChannel(orderId string) bool {
	cp.Lock()
	_, ok := channelMap[orderId]
	if !ok {
		log.Printf("\nChannel creating for order: %s", orderId)
		channelMap[orderId] = make(chan int)
	} else {
		log.Printf("\nChannel already exist %s", orderId)
		return false
	}
	cp.Unlock()
	return true
}

func deleteChannel(orderId string) bool {
	cp.Lock()
	_, ok := channelMap[orderId]
	if ok {
		log.Printf("\nChannel deleting for order: %s", orderId)
		close(channelMap[orderId])
		delete(channelMap, orderId)
	} else {
		log.Printf("\nChannel does not exist %s", orderId)
		return false
	}
	cp.Unlock()
	return true

}

func releaseChannel(orderId string, status int) bool {
	_, ok := channelMap[orderId]
	if ok {
		channelMap[orderId] <- status
		log.Printf("Released channel for order: %s", orderId)
		deleteChannel(orderId)
		return true
	} else {
		log.Printf("Channel does not exist for order: %s", orderId)
		return false
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
