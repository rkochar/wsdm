package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"WDM-G1/shared"
	kafka "github.com/segmentio/kafka-go"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"main/shared"
)

var client *mongo.Client
var ordersCollection *mongo.Collection

func main() {
	go shared.SetUpKafkaListener(
		[]string{"order"}, false,
		func(message *shared.SagaMessage) (*shared.SagaMessage, string) {

			returnMessage := shared.SagaMessageConvertStartToEnd(message)

			if message.Name == "START-UPDATE-ORDER" {
				// ignore error, will not happen
				_, orderID := shared.ConvertStringToMongoID(message.Order.OrderID)

				clientError, serverError := updateOrder(orderID, true)
				if clientError != nil || serverError != nil {
					returnMessage.Name = "ABORT-CHECKOUT-SAGA"
				}

				return returnMessage, "order-ack"
			}

			return nil, ""
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

<<<<<<< HEAD:src/order/app.go
	var err error
	//TODO: implement hash
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://orderdb-svc-0:27017"))
=======
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://orderdb-svc-0:27017"))
>>>>>>> main:order/app.go
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("orders")
	ordersCollection = db.Collection("orders")

	router := mux.NewRouter()
	router.HandleFunc("/orders/create/{user_id}", createOrderHandler)
	router.HandleFunc("/orders/remove/{order_id}", removeOrderHandler)
	router.HandleFunc("/orders/find/{order_id}", findOrderHandler)
	router.HandleFunc("/orders/addItem/{order_id}/{item_id}", addItemHandler)
	router.HandleFunc("/orders/removeItem/{order_id}/{item_id}", removeItemHandler)
	router.HandleFunc("/orders/checkout/{order_id}", checkoutHandler)

<<<<<<< HEAD:src/order/app.go
=======
	router.HandleFunc("/orders/send/{message}", sendKafkaMessageHandler)
	// router.HandleFunc("/orders/kafka/checkout", checkoutKafkaHandler)

>>>>>>> main:order/app.go
	port := os.Getenv("PORT")
	fmt.Printf("Current port is : %s\n", port)
	if port == "" {
		port = "8080"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting order service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

<<<<<<< HEAD:src/order/app.go
func getOrder(orderID *primitive.ObjectID) (error, *shared.Order) {
	filter := bson.M{"_id": orderID}
=======
func getOrder(orderID string) (error, *shared.Order) {
	convertDocIDErr, documentID := shared.ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		return convertDocIDErr, nil
	}

	filter := bson.M{"_id": documentID}
>>>>>>> main:order/app.go
	var order shared.Order
	findDocErr := ordersCollection.FindOne(context.Background(), filter).Decode(&order)
	if findDocErr != nil {
		return findDocErr, nil
	}
	// order.OrderID = orderID.String()
	return nil, &order
}

// Functions only used by http

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
<<<<<<< HEAD:src/order/app.go

	convertUserIDErr, mongoUserID := shared.ConvertStringToMongoID(userID)
	if convertUserIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

=======
	// fmt.Printf("Creating order for user %s\n", userID)

>>>>>>> main:order/app.go
	order := shared.Order{
		Paid:      false,
		Items:     []string{},
		UserID:    mongoUserID.String(),
		TotalCost: 0.0,
	}

	insertResult, mongoInsertErr := ordersCollection.InsertOne(context.Background(), order)
	if mongoInsertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	orderID := insertResult.InsertedID.(primitive.ObjectID).Hex()
<<<<<<< HEAD:src/order/app.go
=======
	// log.Printf("Inserted document with ID: %v\n", insertResult.InsertedID)
>>>>>>> main:order/app.go
	order.OrderID = orderID

	w.Header().Set("Content-Type", "application/json")
	jsonEncodeErr := json.NewEncoder(w).Encode(order)
	if jsonEncodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func removeOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
<<<<<<< HEAD:src/order/app.go

=======
>>>>>>> main:order/app.go
	convertDocIDErr, documentID := shared.ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": documentID}
	_, removeDocErr := ordersCollection.DeleteOne(context.Background(), filter)
	if removeDocErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
<<<<<<< HEAD:src/order/app.go
=======
	// log.Printf("Deleted %d document(s)\n", result.DeletedCount)
	// fmt.Printf("Deleted order with ID: %s\n", orderID)
>>>>>>> main:order/app.go
}

func findOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	convertDocIDErr, documentID := shared.ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	findOrderErr, order := getOrder(documentID)
	order.OrderID = orderID
	if findOrderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonEncodeErr := json.NewEncoder(w).Encode(order)
	if jsonEncodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func addItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]

	// fmt.Printf("Adding item %s to order %s", itemID, orderID)
	// convertItemIDErr, mongoItemID := shared.ConvertStringToMongoID(itemID)
	// if convertItemIDErr != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	// }

	// TODO: use kafka

	getStockResponse, getStockErr := http.Get(fmt.Sprintf("http://localhost:8082/stock/find/%s", itemID))
	// fmt.Printf("response: %s", getStockResponse.StatusCode)
	// fmt.Printf("get stock err: %s", getStockErr)
	if getStockErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer getStockResponse.Body.Close()

	var item shared.Item
	jsonDecodeErr := json.NewDecoder(getStockResponse.Body).Decode(&item)
	if jsonDecodeErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
<<<<<<< HEAD:src/order/app.go

	convertOrderIDErr, mongoOrderID := shared.ConvertStringToMongoID(orderID)
	if convertOrderIDErr != nil {
=======
	// if item.Stock == 0 || item.Price == 0.0 {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	// }

	convertDocIDErr, documentID := shared.ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
>>>>>>> main:order/app.go
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderFilter := bson.M{"_id": mongoOrderID}
	orderUpdate := bson.M{
		"$push": bson.M{
			"items": itemID,
		},
		"$inc": bson.M{
			"totalcost": item.Price,
		},
	}
	_, addItemErr := ordersCollection.UpdateOne(context.Background(), orderFilter, orderUpdate)
	if addItemErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func removeItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]

	convertItemIDErr, mongoItemID := shared.ConvertStringToMongoID(itemID)
	if convertItemIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: use kafka

	getStockResponse, getStockErr := http.Get(fmt.Sprintf("http://localhost:8082/stock/find/%s", mongoItemID))
	if getStockErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer getStockResponse.Body.Close()

	var item shared.Item
	jsonDecodeErr := json.NewDecoder(getStockResponse.Body).Decode(&item)
	if jsonDecodeErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

<<<<<<< HEAD:src/order/app.go
	convertOrderIDErr, mongoOrderID := shared.ConvertStringToMongoID(orderID)
	if convertOrderIDErr != nil {
=======
	convertDocIDErr, documentID := shared.ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
>>>>>>> main:order/app.go
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderFilter := bson.M{"_id": mongoOrderID}
	orderUpdate := bson.M{
		"$pull": bson.M{
			"items": itemID,
		},
		"$inc": bson.M{
			"totalcost": -item.Price,
		},
	}
	_, removeItemErr := ordersCollection.UpdateOne(context.Background(), orderFilter, orderUpdate)
	if removeItemErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

<<<<<<< HEAD:src/order/app.go
=======
// PAYMENT_MAKE_USERID_ORDERID_TOTALCOST
// PAYMENT_CANCEL_USERID_ORDERID_TOTALCOST

func makePayment(order shared.Order) bool {
	URL := fmt.Sprintf("http://localhost:8081/payment/pay/%s/%s/%f", order.UserID, order.OrderID, order.TotalCost)
	fmt.Printf("Making payment via URL: %s\n", URL)
	resp, err := http.Post(URL, "application/json", strings.NewReader(``))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return !(400 <= resp.StatusCode && resp.StatusCode < 500)
}

func subtractStock(itemID string, amount int64) bool {
	URL := fmt.Sprintf("http://localhost:8082/stock/subtract/%s/%d", itemID, amount)
	fmt.Printf("Subtracting stock via URL: %s\n", URL)
	resp, err := http.Post(URL, "application/json", strings.NewReader(``))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return !(400 <= resp.StatusCode && resp.StatusCode < 500)
}

// TODO: set to POST method
>>>>>>> main:order/app.go
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	convertOrderIDErr, mongoOrderID := shared.ConvertStringToMongoID(orderID)
	if convertOrderIDErr != nil {
		fmt.Println("Convert String to Mongo ID error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	getOrderErr, order := getOrder(mongoOrderID)
	order.OrderID = orderID
	fmt.Println("Order ID", orderID, order.OrderID)
	if getOrderErr != nil {
		fmt.Println("Get order error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

<<<<<<< HEAD:src/order/app.go
	//TODO :deepali: create senders and listeners only once and use
	sender := shared.CreateTopicSender("order-ack")
	defer sender.Close()
	fmt.Println("Order ID V2", orderID, order.OrderID)
	message := shared.SagaMessage{
		Name:   "START-CHECKOUT-SAGA",
		SagaID: -1,
		Order:  *order,
	}
	fmt.Println("Message: ", message)
	message.Order.OrderID = orderID
	fmt.Println("Message v2: ", message)
	fmt.Println("Message v3: ", &message)

	sendErr := shared.SendSagaMessage(&message, sender)
	if sendErr != nil {
		fmt.Println("Send Kafka SAGA message error")
		w.WriteHeader(http.StatusInternalServerError)
	}

	// TODO: wait for response and return status
	fmt.Println("TODO TODO TODO")
}

// Functions used only by kafka
=======
	// Step 1: Subtract the stock
	// Compensation: re-add stock
	for _, item := range order.Items {
		// fmt.Printf("Item at index %d: %+v\n", i, item)
		subtractStockSuccess := subtractStock(item, 1)
		if !subtractStockSuccess {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Step 2: Make the payment
	// Compensation: cancel payment
	paymentSuccess := makePayment(*order)
	if !paymentSuccess {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Step 3: Update the order status
	// Compensation: set order to unpaid
	convertDocIDErr, orderDocumentID := shared.ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
>>>>>>> main:order/app.go

func updateOrder(orderID *primitive.ObjectID, status bool) (clientError error, serverError error) {
	orderFilter := bson.M{"_id": orderID}
	orderUpdate := bson.M{
		"$set": bson.M{
			"paid": status,
		},
	}
<<<<<<< HEAD:src/order/app.go
=======
	_, updateOrderErr := ordersCollection.UpdateOne(context.Background(), orderFilter, orderUpdate)
	if updateOrderErr != nil {
		// fmt.Printf("Error during updating of order status: %s", updateOrderErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
>>>>>>> main:order/app.go

	_, clientError = ordersCollection.UpdateOne(context.Background(), orderFilter, orderUpdate)

	return
}

// /orders/checkout/kafka
// func checkoutKafkaHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hi there Kafka!")
// 	fmt.Printf("Doing Kafka checkout test")
//
// 	// to produce messages
// 	topic := "order-ack"
// 	partition := 0
//
// 	sagaID := uuid.New()
//
// 	testItem := Item{
// 		StockID: "a",
// 		Stock:   42,
// 		Price:   69,
// 	}
//
// 	if err := conn.Close(); err != nil {
// 		log.Fatal("failed to close writer:", err)
// 	}
// }
