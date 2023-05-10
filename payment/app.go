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
	"time"
)

type User struct {
	UserID string  `json:"user_id"`
	Credit float64 `json:"credit"`
}

var client *mongo.Client
var paymentCollection *mongo.Collection

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("payment")
	paymentCollection = db.Collection("payment")

	router := mux.NewRouter()
	router.HandleFunc("/payment/pay/{user_id}/{order_id}/{amount}", payHandler)
	router.HandleFunc("/payment/cancel/{user_id}/{order_id}", cancelPaymentHandler)
	router.HandleFunc("/payment/status/{user_id}/{order_id}", paymentStatusHandler)
	router.HandleFunc("/payment/add_funds/{user_id}/{amount}", addFundsHandler)
	router.HandleFunc("/payment/create_user", createUserHandler)
	router.HandleFunc("/payment/find_user/{user_id}", findUserHandler)

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

func payHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]
	amount := vars["amount"]

	fmt.Printf("Subtracting %s from order %s of credit of user %s\n", amount, orderID, userID)
}

func cancelPaymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	fmt.Printf("Cancelling payment of order %s for user %s\n", orderID, userID)
}

func paymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	fmt.Printf("Getting status of payment of order %s for user %s...\n", orderID, userID)
}

func addFundsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	amount := vars["amount"]

	fmt.Printf("Adding %s funds for user %s\n", amount, userID)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Creating a new user with 0 credit...\n")

	user := User{
		Credit: 0.0,
	}

	result, err := paymentCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
	userID := result.InsertedID.(primitive.ObjectID).Hex()
	fmt.Printf("Created a new user with ID: %s\n", userID)
	user.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(user)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
}

func findUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	documentID := ConvertStringToMongoID(userID)

	fmt.Printf("Finding user with ID: %s...\n", userID)

	var result User
	err := paymentCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document found with the given filter")
		} else {
			log.Fatal(err)
		}
		return
	}
	fmt.Printf("Found user: %+v\n", result)
	result.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(result)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
}
