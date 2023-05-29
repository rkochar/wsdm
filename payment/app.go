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

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	Status  string  `json:"status"`
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

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
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

func getUser(userID *primitive.ObjectID) (error, *User) {

	var user User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err, nil
	}
	// fmt.Printf("Found user: %+v\n", user)
	user.UserID = userID.String()
	return nil, &user
}

func getPayment(userID string, orderID string) (error, *Payment) {
	userIdConvErr, mongoUserID := ConvertStringToMongoID(userID)
	if userIdConvErr != nil {
		return userIdConvErr, nil
	}
	orderIdConvErr, mongoOrderID := ConvertStringToMongoID(orderID)
	if orderIdConvErr != nil {
		return orderIdConvErr, nil
	}

	filter := bson.M{"userid": mongoUserID, "orderid": mongoOrderID}
	var payment Payment
	findErr := paymentCollection.FindOne(context.Background(), filter).Decode(&payment)
	if findErr != nil {
		return findErr, nil
	}
	return nil, &payment
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("Yes or no?")
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]
	amount := vars["amount"]
	log.Print("Started stuff")

	userIdConvErr, mongoUserID := ConvertStringToMongoID(userID)
	if userIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("userIdConvErr")
		return
	}
	fmt.Print("Going in")

	orderIdConvErr, mongoOrderID := ConvertStringToMongoID(orderID)
	if orderIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Print("orderIdConvErr")
		return
	}
	fmt.Print("Over here somewhere")
	amountConvErr, amountFloat := ConvertStringToFloat(amount)
	if amountConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Print("amountConvErr")
		return
	}

	fmt.Print("Going in once")

	_, error := onPayment(mongoUserID, mongoOrderID, amountFloat)
	log.Print(error)

	w.Header().Set("Content-Type", "application/json")
	jsonError := json.NewEncoder(w).Encode(error)

	//status1, error1 := onAck(mongoOrderID)

	//if status == Failure {
	//	w.WriteHeader(http.StatusBadRequest)
	//	fmt.Print(error)
	//	return
	//}
	//
	//fmt.Print(error)
	//
	//if status1 == Failure {
	//	w.WriteHeader(http.StatusBadRequest)
	//	fmt.Print(error1)
	//	return
	//}
	//
	//fmt.Print(error1)
	//
	//w.Header().Set("Content-Type", "application/json")
	//jsonErr := json.NewEncoder(w).Encode(jsonError)

	if jsonError != nil {
		fmt.Print("errrorrrr")
		http.Error(w, jsonError.Error(), http.StatusInternalServerError)
		return
	}

}

//TODO: Saga for this: See how to handle this
func cancelPaymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	userIdConvErr, mongoUserID := ConvertStringToMongoID(userID)
	if userIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderIdConvErr, mongoOrderID := ConvertStringToMongoID(orderID)
	if orderIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		getPaymentErr, payment := getPayment(userID, orderID)
		if getPaymentErr != nil {
			return nil, getPaymentErr
		}

		userFilter := bson.M{
			"_id": mongoUserID,
		}
		userUpdate := bson.M{
			"$inc": bson.M{
				"credit": payment.Amount,
			},
		}
		_, userUpdateError := userCollection.UpdateOne(context.Background(), userFilter, userUpdate)
		if userUpdateError != nil {
			return nil, userUpdateError
		}

		paymentFilter := bson.M{
			"userid":  mongoUserID,
			"orderid": mongoOrderID,
		}
		paymentUpdate := bson.M{
			"$set": bson.M{
				"paid": false,
			},
		}
		_, paymentUpdateErr := paymentCollection.UpdateOne(context.Background(), paymentFilter, paymentUpdate)
		if paymentUpdateErr != nil {
			return nil, paymentUpdateErr
		}

		return nil, nil
	}

	session, startSessionErr := client.StartSession()
	// fmt.Printf("Started session")
	if startSessionErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	defer session.EndSession(ctx)
	_, sessionWithTransactionErr := session.WithTransaction(ctx, callback)
	if sessionWithTransactionErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func paymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	orderID := vars["order_id"]

	userIdConvErr, mongoUserID := ConvertStringToMongoID(userID)
	if userIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderIdConvErr, mongoOrderID := ConvertStringToMongoID(orderID)
	if orderIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filter := bson.M{"userid": mongoUserID, "orderid": mongoOrderID}
	var payment Payment
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

	idConvErr, documentID := ConvertStringToMongoID(userID)
	if idConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	amountConvErr, amountFloat := ConvertStringToFloat(amount)
	if amountConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": documentID}
	update := bson.M{
		"$inc": bson.M{
			"credit": +*amountFloat,
		},
	}
	_, updateErr := userCollection.UpdateOne(context.Background(), filter, update)
	response := DoneResponse{}
	if updateErr != nil {
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
	user := User{
		Credit: 0.0,
	}
	log.Print("HE;llllleyyyyyyyy")
	result, insertionError := userCollection.InsertOne(context.Background(), user)
	if insertionError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID := result.InsertedID.(primitive.ObjectID).Hex()
	user.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	jsonError := json.NewEncoder(w).Encode(user)
	if jsonError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func findUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	log.Print("Workssssss")

	userIdConvErr, mongoUserID := ConvertStringToMongoID(userID)
	if userIdConvErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(userID)
		log.Print(userIdConvErr)
		return
	}

	userFindErr, user := getUser(mongoUserID)
	if userFindErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonErr := json.NewEncoder(w).Encode(user)
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func onPayment(userID *primitive.ObjectID, orderId *primitive.ObjectID, amount *float64) (Status, error) {
	log.Print("Atleast here")
	getuserErr, user := getUser(userID)
	if getuserErr != nil {
		log.Print("Get user error")
		log.Print(getuserErr)
		return Failure, getuserErr
	}
	if user.Credit < *amount {
		// fmt.Printf("Not enough stock")
		return Failure, errors.New("not enough credits to pay")
	}

	userFilter := bson.M{
		"_id": userID,
	}
	userUpdate := bson.M{
		"$inc": bson.M{
			"credit": -*amount,
		},
	}

	log.Print("I going to pay")
	_, userUpdateError := userCollection.UpdateOne(context.Background(), userFilter, userUpdate)
	if userUpdateError != nil {
		return Failure, getuserErr
	}

	payment := Payment{
		UserID:  userID.String(),
		OrderID: orderId.String(),
		Amount:  *amount,
		Paid:    true,
		Status:  Pending,
	}
	_, insertErr := paymentCollection.InsertOne(context.Background(), payment)
	if insertErr != nil {
		return Failure, insertErr
	}
	log.Print(insertErr)

	return Completed, nil
}

func onAck(orderDocumentID *primitive.ObjectID) (Status, error) {

	paymentFilter := bson.M{"_id": orderDocumentID}
	paymentUpdate := bson.M{
		"$set": bson.M{
			"status": Completed,
		},
	}
	_, updatePaymentErr := paymentCollection.UpdateOne(context.Background(), paymentFilter, paymentUpdate)

	if updatePaymentErr != nil {
		return Failure, updatePaymentErr
	}
	return Completed, nil
}

func onNack(orderId *primitive.ObjectID, userId *primitive.ObjectID) (Status, error) {
	getPaymentErr, payment := getPayment(userId.String(), orderId.String())
	if getPaymentErr != nil {
		return Rollback_Failure, getPaymentErr
	}

	userFilter := bson.M{
		"_id": orderId,
	}
	userUpdate := bson.M{
		"$inc": bson.M{
			"credit": +payment.Amount,
		},
	}
	_, userUpdateError := userCollection.UpdateOne(context.Background(), userFilter, userUpdate)
	if userUpdateError != nil {
		return Rollback_Failure, userUpdateError
	}

	paymentFilter := bson.M{
		"userid":  userId,
		"orderid": orderId,
	}
	paymentUpdate := bson.M{
		"$set": bson.M{
			"paid":   false,
			"status": Rollback_Completed,
		},
	}
	_, paymentUpdateErr := paymentCollection.UpdateOne(context.Background(), paymentFilter, paymentUpdate)
	if paymentUpdateErr != nil {
		return Rollback_Failure, paymentUpdateErr
	}

	return Rollback_Completed, nil
}
