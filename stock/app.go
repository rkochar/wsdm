package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Item struct {
	StockID string  `json:"item_id"`
	Stock   int64   `json:"stock"`
	Price   float64 `json:"price"`
}

var client *mongo.Client
var stockCollection *mongo.Collection

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("payment")
	stockCollection = db.Collection("stock")

	router := mux.NewRouter()
	router.HandleFunc("/stock/find/{item_id}", findHandler)
	router.HandleFunc("/stock/subtract/{item_id}/{amount}", subtractHandler)
	router.HandleFunc("/stock/add/{item_id}/{amount}", addHandler)
	router.HandleFunc("/stock/item/create/{price}", createHandler)

	port := os.Getenv("PORT")
	fmt.Printf("Current port is: %s\n", port)
	if port == "" {
		port = "8080"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting stock service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func getItem(itemID string) Item {
	documentID := ConvertStringToMongoID(itemID)

	var item Item
	err := stockCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No item found with the given filter")
		} else {
			log.Fatal(err)
		}
		return item
	}
	fmt.Printf("Found item: %+v\n", item)
	item.StockID = itemID
	return item
}

func updateItemStock(item Item) bool {
	documentID := ConvertStringToMongoID(item.StockID)
	update := bson.M{
		"$set": bson.M{
			"stock": item.Stock,
		},
	}
	_, updateErr := stockCollection.UpdateOne(context.Background(), bson.M{"_id": documentID}, update)
	if updateErr != nil {
		log.Fatal(updateErr)
		return false
	}
	return true
}

func findHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]

	fmt.Printf("Find: %s\n", itemID)
	item := getItem(itemID)

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(item)
	if err1 != nil {
		log.Fatal(err1)
	}
}

func subtractHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]
	amount := vars["amount"]
	intAmount := ConvertStringToInt(amount)

	fmt.Printf("Subtracting %s from order %s\n", amount, itemID)
	item := getItem(itemID)
	item.Stock -= intAmount
	success := updateItemStock(item)
	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]
	amount := vars["amount"]
	intAmount := ConvertStringToInt(vars["amount"])

	fmt.Printf("Adding %s to order %s\n", amount, itemID)
	item := getItem(itemID)
	item.Stock += intAmount
	success := updateItemStock(item)
	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	price := vars["price"]
	priceFloat, err := strconv.ParseFloat(price, 64)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Creating item with price %s\n", price)
	stock := Item{
		Stock: 0,
		Price: priceFloat,
	}
	result, insertErr := stockCollection.InsertOne(context.Background(), stock)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	stockID := result.InsertedID.(primitive.ObjectID).Hex()
	fmt.Printf("Created a new item with ID: %s\n", stockID)
	stock.StockID = stockID

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(stock)
	if err1 != nil {
		log.Fatal(err1)
	}
}
