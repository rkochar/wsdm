package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"time"

	"WDM-G1/shared"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var clients [shared.NUM_DBS]*mongo.Client
var ordersCollections [shared.NUM_DBS]*mongo.Collection

//var client *mongo.Client
//var ordersCollection *mongo.Collection

const partition = 1

func main() {
	go shared.SetUpKafkaListener(
		[]string{"order"}, false,
		func(message *shared.SagaMessage) (*shared.SagaMessage, string) {

			returnMessage := shared.SagaMessageConvertStartToEnd(message)

			if message.Name == "START-UPDATE-ORDER" {
				// ignore error, will not happen
				_, orderID := shared.ConvertStringToUUID(message.Order.OrderID)

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

	setupErr := setupDBConnections(ctx)
	if setupErr != nil {
		log.Fatal(setupErr)
	}
	for i := 0; i < shared.NUM_DBS; i++ {
		defer clients[i].Disconnect(ctx)
	}

	router := mux.NewRouter()
	router.HandleFunc("/orders/create/{user_id}", createOrderHandler)
	router.HandleFunc("/orders/remove/{order_id}", removeOrderHandler)
	router.HandleFunc("/orders/find/{order_id}", findOrderHandler)
	router.HandleFunc("/orders/addItem/{order_id}/{item_id}", addItemHandler)
	router.HandleFunc("/orders/removeItem/{order_id}/{item_id}", removeItemHandler)
	router.HandleFunc("/orders/checkout/{order_id}", checkoutHandler)

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

func setupDBConnections(ctx context.Context) error {
	for i := 0; i < shared.NUM_DBS; i++ {
		mongoURL := fmt.Sprintf("mongodb://orderdb-service-%d:27017", i)
		//mongoURL := "mongodb://localhost:27017"
		fmt.Printf("%d MongoDB URL: %s", i, mongoURL)
		var err error
		var client *mongo.Client
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))

		if err != nil {
			return err
		}
		clients[i] = client
		ordersCollections[i] = client.Database("orders").Collection("orders")
	}
	return nil
}

func getOrder(orderID *uuid.UUID) (error, *shared.Order) {
	ordersCollection := getOrdersCollection(*orderID)
	filter := bson.M{"_id": orderID}
	var order shared.Order
	findDocErr := ordersCollection.FindOne(context.Background(), filter).Decode(&order)
	if findDocErr != nil {
		return findDocErr, nil
	}
	return nil, &order
}

func getOrdersCollection(orderID uuid.UUID) *mongo.Collection {
	databaseNum := shared.HashUUID(orderID)
	return ordersCollections[databaseNum]
}

// Functions only used by http

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	convertUserIDErr, mongoUserID := shared.ConvertStringToUUID(userID)
	if convertUserIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderID := shared.GetNewID()

	order := shared.Order{
		ID:        orderID,
		OrderID:   orderID.String(),
		Paid:      false,
		Items:     []string{},
		UserID:    mongoUserID.String(),
		TotalCost: 0.0,
	}

	ordersCollection := getOrdersCollection(orderID)
	_, mongoInsertErr := ordersCollection.InsertOne(context.Background(), order)
	if mongoInsertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonEncodeErr := json.NewEncoder(w).Encode(order)
	if jsonEncodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func removeOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	convertDocIDErr, documentID := shared.ConvertStringToUUID(orderID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ordersCollection := getOrdersCollection(*documentID)
	filter := bson.M{"_id": documentID}
	_, removeDocErr := ordersCollection.DeleteOne(context.Background(), filter)
	if removeDocErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func findOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	convertDocIDErr, documentID := shared.ConvertStringToUUID(orderID)
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

	log.Printf("Adding item %s to order %s", itemID, orderID)

	// fmt.Printf("Adding item %s to order %s", itemID, orderID)
	convertItemIDErr, mongoItemID := shared.ConvertStringToUUID(itemID)
	if convertItemIDErr != nil {
		log.Print(convertItemIDErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: use kafka

	stockURL := fmt.Sprintf("http://stock-service:5000/stock/find/%s", mongoItemID.String())
	//stockURL := fmt.Sprintf("http://localhost:8082/stock/find/%s", mongoItemID.String())
	getStockResponse, getStockErr := http.Get(stockURL)
	log.Printf("response: %s", getStockResponse.StatusCode)
	log.Printf("get stock err: %s", getStockErr)
	if getStockErr != nil {
		log.Print(getStockResponse)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer getStockResponse.Body.Close()

	var item shared.Item
	jsonDecodeErr := json.NewDecoder(getStockResponse.Body).Decode(&item)
	log.Printf("json body err: %s", item)
	if jsonDecodeErr != nil {
		log.Print(jsonDecodeErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	convertOrderIDErr, mongoOrderID := shared.ConvertStringToUUID(orderID)
	if convertOrderIDErr != nil {
		log.Print(jsonDecodeErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ordersCollection := getOrdersCollection(*mongoOrderID)
	orderFilter := bson.M{"_id": mongoOrderID}
	orderUpdate := bson.M{
		"$push": bson.M{
			"items": mongoItemID.String(),
		},
		"$inc": bson.M{
			"totalcost": item.Price,
		},
	}
	result := shared.UpdateRecord(ordersCollection, orderFilter, orderUpdate)
	if result.Err() != nil {
		log.Print(result.Err())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func removeItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]

	convertItemIDErr, mongoItemID := shared.ConvertStringToUUID(itemID)
	if convertItemIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: use kafka
	stockURL := fmt.Sprintf("http://stock-service:5000/find/%s", mongoItemID.String())
	//stockURL := fmt.Sprintf("http://localhost:8082/find/%s", mongoItemID.String())
	getStockResponse, getStockErr := http.Get(stockURL)
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

	convertOrderIDErr, mongoOrderID := shared.ConvertStringToUUID(orderID)
	if convertOrderIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ordersCollection := getOrdersCollection(*mongoOrderID)
	orderFilter := bson.M{"_id": mongoOrderID}
	orderUpdate := bson.M{
		"$pull": bson.M{
			"items": itemID,
		},
		"$inc": bson.M{
			"totalcost": -item.Price,
		},
	}
	result := shared.UpdateRecord(ordersCollection, orderFilter, orderUpdate)
	if result.Err() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	convertOrderIDErr, mongoOrderID := shared.ConvertStringToUUID(orderID)
	if convertOrderIDErr != nil {
		log.Println("Convert String to Mongo ID error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	getOrderErr, order := getOrder(mongoOrderID)
	order.OrderID = orderID
	// fmt.Println("Order ID", orderID, order.OrderID)
	if getOrderErr != nil {
		log.Println("Get order error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sender := shared.CreateConnection("order-ack", partition)
	defer sender.Close()

	message := shared.SagaMessage{
		Name:   "START-CHECKOUT-SAGA",
		SagaID: -1,
		Order:  *order,
	}
	// message.Order.OrderID = orderID

	sendErr := shared.SendSagaMessage(&message, sender)
	if sendErr != nil {
		log.Println("Send Kafka SAGA message error")
		w.WriteHeader(http.StatusInternalServerError)
	}

	// TODO: wait for response and return status
	log.Println("TODO TODO TODO")
}

// Functions used only by kafka

func updateOrder(orderID *uuid.UUID, status bool) (clientError error, serverError error) {
	ordersCollection := getOrdersCollection(*orderID)
	orderFilter := bson.M{"_id": orderID}
	orderUpdate := bson.M{
		"$set": bson.M{
			"paid": status,
		},
	}
	result := shared.UpdateRecord(ordersCollection, orderFilter, orderUpdate)
	if result.Err() != nil {
		log.Print(result.Err())
		return nil, result.Err()
	}
	return nil, nil
}
