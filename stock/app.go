package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type FindResponse struct {
	ItemID string `json:"item_id"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/stock/find/{item_id}/", findHandler)

	// Set the listening address and port for the server
	addr := ":8080"
	fmt.Printf("Starting server at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func findHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]

	response := FindResponse{
		ItemID: itemID,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
