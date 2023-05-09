package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/order/create/{order_id}/", createOrderHandler)

	// Set the listening address and port for the server
	addr := ":8080"
	fmt.Printf("Starting order service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	response := CreateOrderResponse{
		OrderID: orderID,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
