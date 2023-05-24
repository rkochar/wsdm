package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	kafka "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	topic  = "order-ack"
	broker = "kafka-service.kafka.svc.cluster.local:9092"
	//broker1Address = "localhost:9093"
	//broker2Address = "localhost:9094"
	//broker3Address = "localhost:9095"
)

type Order struct {
	OrderID   string   `json:"order_id"`
	Paid      bool     `json:"paid"`
	Items     []string `json:"items"`
	UserID    string   `json:"user_id"`
	TotalCost float64  `json:"total_cost"`
}

type Item struct {
	StockID string  `json:"item_id"`
	Stock   int64   `json:"stock"`
	Price   float64 `json:"price"`
}

var client *mongo.Client
var ordersCollection *mongo.Collection

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go produce()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-stock:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	db := client.Database("orders")
	ordersCollection = db.Collection("orders")

	router := mux.NewRouter()
	router.HandleFunc("/", greetingHandler)
	router.HandleFunc("/orders/create/{user_id}", createOrderHandler)
	router.HandleFunc("/orders/remove/{order_id}", removeOrderHandler)
	router.HandleFunc("/orders/find/{order_id}", findOrderHandler)
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
	fmt.Printf("Starting order service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func getOrder(orderID string) Order {
	documentID := ConvertStringToMongoID(orderID)

	var order Order
	err := ordersCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No orders found with the given filter")
		} else {
			log.Fatal(err)
		}
		return order
	}
	order.OrderID = orderID
	return order
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!Welcome to order")
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

	order := getOrder(orderID)
	w.Header().Set("Content-Type", "application/json")
	err1 := json.NewEncoder(w).Encode(order)
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

	resp, err := http.Get(fmt.Sprintf("http://localhost:8082/stock/find/%s", itemID))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var item Item
	err = json.NewDecoder(resp.Body).Decode(&item)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Received response: %+v\n", item)
	if item.Stock == 0 || item.Price == 0.0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("Adding item %s to order %s", itemID, orderID)
	documentID := ConvertStringToMongoID(orderID)
	order := getOrder(orderID)
	fmt.Printf("Found order: %+v\n", order)
	update := bson.M{
		"$push": bson.M{
			"items": itemID,
		},
		"$inc": bson.M{
			"totalcost": item.Price,
		},
	}
	_, addErr := ordersCollection.UpdateOne(context.Background(), bson.M{"_id": documentID}, update)
	if addErr != nil {
		log.Fatal(addErr)
		return
	}
}

// TODO: set to DELETE method
func removeItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]
	fmt.Printf("Removing item %s from order %s\n", itemID, orderID)
	documentID := ConvertStringToMongoID(orderID)

	order := getOrder(orderID)
	fmt.Printf("Found order: %+v\n", order)
	update := bson.M{"$pull": bson.M{"items": itemID}}
	_, removeErr := ordersCollection.UpdateOne(context.Background(), bson.M{"_id": documentID}, update)
	if removeErr != nil {
		log.Fatal(removeErr)
		return
	}
}

func makePayment(order Order) bool {
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
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	fmt.Printf("Checkout for order %s\n", orderID)
	order := getOrder(orderID)
	fmt.Printf("Found the order: %+v\n", order)

	// Step 1: Make the payment
	paymentSuccess := makePayment(order)
	if !paymentSuccess {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Step 2: Subtract the stock
	for i, item := range order.Items {
		fmt.Printf("Item at index %d: %+v\n", i, item)
		subtractStockSuccess := subtractStock(item, 1)
		if !subtractStockSuccess {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// Step 3: Update the order status
	update := bson.M{"$set": bson.M{"paid": true}}
	orderDocumentID := ConvertStringToMongoID(orderID)
	_, updateOrderErr := ordersCollection.UpdateOne(context.Background(), bson.M{"_id": orderDocumentID}, update)
	if updateOrderErr != nil {
		fmt.Printf("Error during updating of order status: %s", updateOrderErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Step 4: Return checkout status
	w.WriteHeader(http.StatusAccepted)
}

func produce() {
	topic := topic
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", broker, topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte("one!")},
		kafka.Message{Value: []byte("two!")},
		kafka.Message{Value: []byte("three!")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
