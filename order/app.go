package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Order struct {
	OrderID   string   `json:"order_id"`
	Paid      bool     `json:"paid"`
	Items     []string `json:"items"`
	UserID    string   `json:"user_id"`
	TotalCost float64  `json:"total_cost"`
}

var client *mongo.Client
var ordersCollection *mongo.Collection

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("orders")
	ordersCollection = db.Collection("orders")

	router := mux.NewRouter()
	router.HandleFunc("/orders/create/{user_id}/", createOrderHandler)
	router.HandleFunc("/orders/remove/{order_id}/", removeOrderHandler)
	router.HandleFunc("/orders/find/{order_id}/", findOrderHandler)
	router.HandleFunc("/orders/addItem/{order_id}/{item_id}", addItemHandler)
	router.HandleFunc("/orders/removeItem/{order_id}/{item_id}", removeItemHandler)
	router.HandleFunc("/orders/checkout/{order_id}", checkoutHandler)

	port := os.Getenv("PORT")
	fmt.Printf("Current port is: %s\n", port)
	if port == "" {
		port = "8080"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting server at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

// TODO: set to POST method
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	/*
		Handler for the route responsible for creating a new order given a certain user ID.
	*/
	vars := mux.Vars(r)
	userID := vars["user_id"]
	fmt.Printf("Creating order for user %s\n", userID)

	order := Order{
		Paid:      false,
		Items:     []string{},
		UserID:    userID,
		TotalCost: 0.0,
	}

	result, err1 := ordersCollection.InsertOne(context.Background(), order)
	if err1 != nil {
		log.Fatal(err1)
	}
	orderID := result.InsertedID.(primitive.ObjectID).Hex()
	log.Printf("Inserted document with ID: %v\n", result.InsertedID)
	order.OrderID = orderID

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// TODO: set to DELETE method
func removeOrderHandler(w http.ResponseWriter, r *http.Request) {
	/*
		Handler for the route responsible for deleting an order given its ID.
	*/
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	documentID := ConvertStringToMongoID(orderID)

	result, err := ordersCollection.DeleteOne(context.Background(), bson.M{"_id": documentID})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Deleted %d document(s)\n", result.DeletedCount)
	fmt.Printf("Deleted order with ID: %s\n", orderID)
}

// TODO: set to GET method
func findOrderHandler(w http.ResponseWriter, r *http.Request) {
	/*
		Handler for the route responsible for finding and returning an order.
	*/
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	documentID := ConvertStringToMongoID(orderID)

	var result Order
	err := ordersCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document found with the given filter")
		} else {
			log.Fatal(err)
		}
		return
	}
	fmt.Printf("Found document: %+v\n", result)
	result.OrderID = orderID

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(result)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
}

// TODO: set to POST method
func addItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]
	fmt.Printf("Adding item %s to order %s", itemID, orderID)
	documentID := ConvertStringToMongoID(orderID)

	var result Order
	err := ordersCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document found with the given filter")
		} else {
			log.Fatal(err)
		}
		return
	}
	fmt.Printf("Found document: %+v\n", result)

	update := bson.M{"$push": bson.M{"items": itemID}}
	_, addErr := ordersCollection.UpdateOne(context.Background(), bson.M{"_id": documentID}, update)
	if addErr != nil {
		log.Fatal(addErr)
		return
	}
	result.Items = append(result.Items, itemID)
	fmt.Printf("Document with added items: %+v\n", result)
}

// TODO: set to DELETE method
func removeItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]
	fmt.Printf("Removing item %s from order %s", itemID, orderID)
	documentID := ConvertStringToMongoID(orderID)

	var result Order
	err := ordersCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document found with the given filter")
		} else {
			log.Fatal(err)
		}
		return
	}
	fmt.Printf("Found document: %+v\n", result)

	update := bson.M{"$pull": bson.M{"items": itemID}}
	_, removeErr := ordersCollection.UpdateOne(context.Background(), bson.M{"_id": documentID}, update)
	if removeErr != nil {
		log.Fatal(removeErr)
	}
}

// TODO: implement this method
// TODO: set to POST method
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	fmt.Printf("Checkout for order %s", orderID)
}
