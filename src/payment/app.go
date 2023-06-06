package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"main/shared"
)

type DoneResponse struct {
	Done bool `json:"done"`
}

type PaidResponse struct {
	Paid bool `json:"paid"`
}

var clients [shared.NUM_DBS]*mongo.Client
var userCollections [shared.NUM_DBS]*mongo.Collection
var paymentCollections [shared.NUM_DBS]*mongo.Collection

// var client *mongo.Client
// var userCollection *mongo.Collection
// var paymentCollection *mongo.Collection

func main() {
	go shared.SetUpKafkaListener(
		[]string{"payment"}, false,
		func(message *shared.SagaMessage) (*shared.SagaMessage, string) {
			returnMessage := shared.SagaMessageConvertStartToEnd(message)

			// TODO: remove code duplication

			if message.Name == "START-MAKE-PAYMENT" {
				// ignore error, wil not happen
				_, mongoUserID := shared.ConvertStringToUUID(message.Order.UserID)
				_, mongoOrderID := shared.ConvertStringToUUID(message.Order.OrderID)

				clientError, serverError := pay(mongoUserID, mongoOrderID, &message.Order.TotalCost)
				if clientError != nil || serverError != nil {
					log.Print(clientError, serverError)
					returnMessage.Name = "ABORT-CHECKOUT-SAGA"
				}
				return returnMessage, "payment-ack"
			}

			if message.Name == "START-CANCEL-PAYMENT" {
				// ignore error, wil not happen
				_, mongoUserID := shared.ConvertStringToUUID(message.Order.UserID)
				_, mongoOrderID := shared.ConvertStringToUUID(message.Order.OrderID)

				clientError, serverError := cancelPayment(mongoUserID, mongoOrderID)
				if clientError != nil || serverError != nil {
					returnMessage.Name = "ABORT-CHECKOUT-SAGA"
				}

				return returnMessage, "payment-ack"
			}

			return nil, ""
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	setupErr := setupDBConnections(ctx)
	if setupErr != nil {
		log.Fatal(setupErr)
	}
	for i := 0; i < shared.NUM_DBS; i++ {
		defer clients[i].Disconnect(ctx)
	}

	router := mux.NewRouter()
	router.HandleFunc("/pay/{user_id}/{order_id}/{amount}", payHandler)
	router.HandleFunc("/cancel/{user_id}/{order_id}", cancelPaymentHandler)
	router.HandleFunc("/status/{user_id}/{order_id}", paymentStatusHandler)
	router.HandleFunc("/add_funds/{user_id}/{amount}", addFundsHandler)
	router.HandleFunc("/create_user", createUserHandler)
	router.HandleFunc("/find_user/{user_id}", findUserHandler)
	router.HandleFunc("/", greetingHandler)

	port := os.Getenv("PORT")
	fmt.Printf("Current port is : %s\n", port)
	if port == "" {
		port = "8081"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting payment service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func setupDBConnections(ctx context.Context) error {
	for i := 0; i < shared.NUM_DBS; i++ {
		mongoURL := fmt.Sprintf("mongodb://paymentdb-service-%d:27017", i)
		fmt.Printf("%d MongoDB URL: %s\n", i, mongoURL)
		var err error
		var client *mongo.Client
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))

		if err != nil {
			return err
		}
		clients[i] = client
		userCollections[i] = client.Database("payment").Collection("users")
		paymentCollections[i] = client.Database("payment").Collection("payments")
	}
	return nil
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("URL Path:", r.URL.Path)

	// You can also print the full URL if the request included it
	if r.URL.RawQuery != "" {
		fmt.Println("Full URL:", r.URL.String())
	}
	log.Print("Hello welcome to paymnt!!")
}

func getUser(documentID *uuid.UUID) (error, *shared.User) {
	userCollection := getUserCollection(documentID)

	var user shared.User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&user)
	if err != nil {
		return err, nil
	}
	user.ID = *documentID
	user.UserID = documentID.String()
	return nil, &user
}

func getPayment(userID *uuid.UUID, orderID *uuid.UUID) (error, *shared.Payment) {
	paymentCollection := getPaymentCollection(userID, orderID)

	filter := bson.M{"userid": userID, "orderid": orderID}
	var payment shared.Payment
	findErr := paymentCollection.FindOne(context.Background(), filter).Decode(&payment)
	if findErr != nil {
		return findErr, nil
	}
	return nil, &payment
}

func getUserCollection(userID *uuid.UUID) *mongo.Collection {
	databaseNum := shared.HashUUID(*userID)
	return userCollections[databaseNum]
}

func getPaymentCollection(userID *uuid.UUID, orderID *uuid.UUID) *mongo.Collection {
	databaseNum := shared.HashTwoUUIDs(*userID, *orderID)
	return paymentCollections[databaseNum]
}

// Functions only used by http

func paymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	userIdConvErr, mongoUserID := shared.ConvertStringToUUID(userID)
	if userIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderIdConvErr, mongoOrderID := shared.ConvertStringToUUID(orderID)
	if orderIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	paymentCollection := getPaymentCollection(mongoUserID, mongoOrderID)
	filter := bson.M{"userid": mongoUserID, "orderid": mongoOrderID}
	var payment shared.Payment
	findErr := paymentCollection.FindOne(context.Background(), filter).Decode(&payment)
	if findErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := PaidResponse{
		Paid: payment.Paid,
	}
	w.Header().Set("Content-Type", "application/json")
	jsonErr := json.NewEncoder(w).Encode(response)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}
}

func addFundsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	amount := vars["amount"]

	idConvErr, documentID := shared.ConvertStringToUUID(userID)
	if idConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	amountConvErr, amountInt := shared.ConvertStringToInt(amount)
	if amountConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userCollection := getUserCollection(documentID)
	filter := bson.M{"_id": documentID}
	update := bson.M{
		"$inc": bson.M{
			"credit": amountInt,
		},
	}
	result := shared.UpdateRecord(userCollection, filter, update)
	response := DoneResponse{}
	if result.Err() != nil {
		response.Done = false
	} else {
		response.Done = true
	}

	w.Header().Set("Content-Type", "application/json")
	jsonErr := json.NewEncoder(w).Encode(response)
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("creates user handler")

	user := shared.User{
		Credit: 0.0,
	}
	userID := shared.GetNewID()
	user.ID = userID
	user.UserID = userID.String()
	userCollection := getUserCollection(&userID)
	fmt.Printf("New user: %+v\n", user)
	_, insertionError := userCollection.InsertOne(context.Background(), user)
	if insertionError != nil {
		log.Print(insertionError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonError := json.NewEncoder(w).Encode(user)
	if jsonError != nil {
		log.Print(jsonError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func findUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	userIdConvErr, mongoUserID := shared.ConvertStringToUUID(userID)
	if userIdConvErr != nil {
		log.Print(userIdConvErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userFindErr, user := getUser(mongoUserID)
	if userFindErr != nil {
		log.Print(userFindErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonErr := json.NewEncoder(w).Encode(user)
	if jsonErr != nil {
		log.Print(jsonErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Functions used by http and kafka

func payHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]
	amount := vars["amount"]

	userIdConvErr, mongoUserID := shared.ConvertStringToUUID(userID)
	if userIdConvErr != nil {
		log.Print(userIdConvErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderIdConvErr, mongoOrderID := shared.ConvertStringToUUID(orderID)
	if orderIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	amountConvErr, amountInt := shared.ConvertStringToInt(amount)
	if amountConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientError, serverError := pay(mongoUserID, mongoOrderID, amountInt)

	if clientError != nil {
		log.Print(clientError)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if serverError != nil {
		log.Print(serverError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func pay(userID *uuid.UUID, orderID *uuid.UUID, amount *int64) (clientError error, serverError error) {
	getuserErr, user := getUser(userID)
	if getuserErr != nil {
		log.Print("user not found")
		clientError = getuserErr
		return
	}
	if user.Credit < *amount {
		log.Print("not enough credits to pay")
		clientError = errors.New("not enough credits to pay")
		return
	}

	userFilter := bson.M{
		"_id": userID,
	}
	userUpdate := bson.M{
		"$inc": bson.M{
			"credit": -*amount,
		},
	}

	userCollection := getUserCollection(userID)
	result := shared.UpdateRecord(userCollection, userFilter, userUpdate)
	if result.Err() != nil {
		serverError = result.Err()
		return
	}

	paymentCollection := getPaymentCollection(userID, orderID)
	payment := shared.Payment{
		UserID:  userID.String(),
		OrderID: orderID.String(),
		Amount:  *amount,
		Paid:    true,
	}
	_, insertErr := paymentCollection.InsertOne(context.Background(), payment)
	if insertErr != nil {
		serverError = insertErr
		return
	}

	return
}

func cancelPaymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	// TODO: send kafka message to cancel order
	userIdConvErr, mongoUserID := shared.ConvertStringToUUID(userID)
	if userIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderIdConvErr, mongoOrderID := shared.ConvertStringToUUID(orderID)
	if orderIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientError, serverError := cancelPayment(mongoUserID, mongoOrderID)
	if clientError != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if serverError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func cancelPayment(userID *uuid.UUID, orderID *uuid.UUID) (clientError error, serverError error) {
	getPaymentErr, payment := getPayment(userID, orderID)
	if getPaymentErr != nil {
		clientError = getPaymentErr
		return
	}

	userCollection := getUserCollection(userID)
	userFilter := bson.M{
		"_id": userID,
	}
	userUpdate := bson.M{
		"$inc": bson.M{
			"credit": payment.Amount,
		},
	}
	result := shared.UpdateRecord(userCollection, userFilter, userUpdate)
	if result.Err() != nil {
		serverError = result.Err()
		return
	}

	paymentCollection := getPaymentCollection(userID, orderID)
	paymentFilter := bson.M{
		"userid":  userID,
		"orderid": orderID,
	}
	paymentUpdate := bson.M{
		"$set": bson.M{
			"paid": false,
		},
	}
	result = shared.UpdateRecord(paymentCollection, paymentFilter, paymentUpdate)
	if result.Err() != nil {
		serverError = result.Err()
	}
	return
}
