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

type User struct {
	UserID string  `json:"user_id"`
	Credit float64 `json:"credit"`
}

type Payment struct {
	UserID  string  `json:"user_id"`
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Paid    bool    `json:"paid"`
}

type DoneResponse struct {
	Done bool `json:"done"`
}

type PaidResponse struct {
	Paid bool `json:"paid"`
}

var client *mongo.Client
var userCollection *mongo.Collection
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
	userCollection = db.Collection("users")
	paymentCollection = db.Collection("payments")

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
		port = "8081"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting payment service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func getUser(userID string) User {
	documentID := ConvertStringToMongoID(userID)

	var user User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No user found with the given filter")
		} else {
			log.Fatal(err)
		}
		return user
	}
	fmt.Printf("Found user: %+v\n", user)
	user.UserID = userID
	return user
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]
	amount := vars["amount"]
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Subtracting %s from order %s of credit of user %s\n", amount, orderID, userID)
	user := getUser(userID)
	if amountFloat > user.Credit {
		fmt.Printf("User has insufficient credit...")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user.Credit -= amountFloat
	payment := Payment{
		UserID:  userID,
		OrderID: orderID,
		Amount:  amountFloat,
		Paid:    true,
	}
	_, insertErr := paymentCollection.InsertOne(context.Background(), payment)
	if insertErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(insertErr)
	}

	userDocumentID := ConvertStringToMongoID(userID)
	update := bson.M{
		"$set": bson.M{
			"credit": user.Credit,
		},
	}
	_, updateErr := userCollection.UpdateOne(context.Background(), bson.M{"_id": userDocumentID}, update)
	if updateErr != nil {
		log.Fatal(updateErr)
	}
}

func cancelPaymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	fmt.Printf("Cancelling payment of order %s for user %s\n", orderID, userID)

	filter := bson.M{"userid": userID, "orderid": orderID}
	var payment Payment
	findErr := paymentCollection.FindOne(context.Background(), filter).Decode(&payment)
	if findErr != nil {
		log.Fatal(findErr)
	}

	user := getUser(userID)
	user.Credit += payment.Amount
	payment.Paid = false

	userCreditUpdate := bson.M{
		"$set": bson.M{
			"credit": user.Credit,
		},
	}
	userDocumentID := ConvertStringToMongoID(userID)
	userCollection.UpdateOne(context.Background(), bson.M{"_id": userDocumentID}, userCreditUpdate)

	paymentUpdate := bson.M{
		"$set": bson.M{
			"paid": false,
		},
	}
	paymentFilter := bson.M{
		"userid":  payment.UserID,
		"orderID": payment.OrderID,
	}
	paymentCollection.UpdateOne(context.Background(), paymentFilter, paymentUpdate)
}

func paymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	fmt.Printf("Getting status of payment of order %s for user %s...\n", orderID, userID)
	filter := bson.M{"userid": userID, "orderid": orderID}
	var payment Payment
	findErr := paymentCollection.FindOne(context.Background(), filter).Decode(&payment)
	if findErr != nil {
		log.Fatal(findErr)
	}

	response := PaidResponse{
		Paid: payment.Paid,
	}
	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(response)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
}

func addFundsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	amount := vars["amount"]
	documentID := ConvertStringToMongoID(userID)

	fmt.Printf("Adding %s funds for user %s\n", amount, userID)
	user := getUser(userID)

	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		fmt.Println(err)
	}
	user.Credit += amountFloat
	response := DoneResponse{}

	update := bson.M{
		"$set": bson.M{
			"credit": user.Credit,
		},
	}
	_, updateErr := userCollection.UpdateOne(context.Background(), bson.M{"_id": documentID}, update)
	if updateErr != nil {
		//log.Fatal(addErr)
		response.Done = false
	} else {
		response.Done = true
	}

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(response)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Creating a new user with 0 credit...\n")

	user := User{
		Credit: 0.0,
	}
	result, err := userCollection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
	userID := result.InsertedID.(primitive.ObjectID).Hex()
	fmt.Printf("Created a new user with ID: %s\n", userID)
	user.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(user)
	if err1 != nil {
		log.Fatal(err1)
		//http.Error(w, err1.Error(), http.StatusInternalServerError)
		//return
	}
}

func findUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	fmt.Printf("Finding user with ID: %s...\n", userID)
	user := getUser(userID)

	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(user)
	if err1 != nil {
		http.Error(w, err1.Error(), http.StatusInternalServerError)
		return
	}
}
