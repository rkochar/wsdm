package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Stock struct {
	itemID string
	stock  int
	price  float32
}

var stock = make(map[string]Stock)

type FindResponse struct {
	Stock int     `json:"stock"`
	Price float32 `json:"price"`
}

func findHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	itemID := vars["item_id"]
	item, ok := stock[itemID]

	if !ok {
		http.Error(responseWriter, "unable to find item", http.StatusBadRequest)
		return
	}

	response := FindResponse{
		Stock: item.stock,
		Price: item.price,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(responseWriter).Encode(response)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

type CountMode int

const (
	Add CountMode = iota
	Subtract
)

func subtractHandler(responseWriter http.ResponseWriter, request *http.Request) {
	countHandler(responseWriter, request, Subtract)
}

func addHandler(responseWriter http.ResponseWriter, request *http.Request) {
	countHandler(responseWriter, request, Add)
}

func countHandler(responseWriter http.ResponseWriter, request *http.Request, mode CountMode) {
	vars := mux.Vars(request)

	amount, err := strconv.ParseInt(vars["amount"], 10, 32)
	if err != nil {
		http.Error(responseWriter, "unable to convert amount to int", http.StatusBadRequest)
		return
	}

	itemID := vars["item_id"]
	oldItem, ok := stock[itemID]

	if !ok {
		http.Error(responseWriter, "unable to find item", http.StatusBadRequest)
		return
	}

	if mode == Subtract {
		amount *= -1
	}

	newAmount := oldItem.stock + int(amount)
	if newAmount < 0 {
		http.Error(responseWriter, "unable to subtract stock (not enough)", http.StatusBadRequest)
		return
	}

	stock[itemID] = Stock{
		itemID: oldItem.itemID,
		stock:  oldItem.stock + int(amount),
		price:  oldItem.price,
	}
}

type CreateResponse struct {
	ItemID string `json:"item_id"`
}

func createHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	price, err := strconv.ParseFloat(vars["price"], 32)

	if err != nil {
		http.Error(responseWriter, "unable to convert price to float", http.StatusBadRequest)
		return
	}

	newID := uuid.New().String()
	stock[newID] = Stock{
		itemID: newID,
		stock:  0,
		price:  float32(price),
	}

	response := CreateResponse{
		ItemID: newID,
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(responseWriter).Encode(response)

	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/stock/find/{item_id}", findHandler)
	router.HandleFunc("/stock/subtract/{item_id}/{amount}", subtractHandler)
	router.HandleFunc("/stock/add/{item_id}/{amount}", addHandler)
	router.HandleFunc("/stock/item/create/{price}", createHandler)

	// Set the listening address and port for the server
	addr := ":8080"
	fmt.Printf("Starting stock service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
