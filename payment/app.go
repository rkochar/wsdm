package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type PayResponse struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
	Amount  string `json:"amount"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/payment/pay/{user_id}/{order_id}/{amount}", payHandler)

	// Set the listening address and port for the server
	addr := ":8080"
	fmt.Printf("Starting payment service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]
	amount := vars["amount"]

	response := PayResponse{
		OrderID: orderID,
		UserID:  userID,
		Amount:  amount,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
