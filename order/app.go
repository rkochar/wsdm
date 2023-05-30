package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	//"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
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

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://stockdb-svc-0:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("orders")
	ordersCollection = db.Collection("orders")

	// Run the web server.
	log.Println("start producer-api ... !!")

	router := mux.NewRouter()
	router.HandleFunc("/", greetingHandler)
	router.HandleFunc("/orders/create/{user_id}", createOrderHandler)
	router.HandleFunc("/orders/remove/{order_id}", removeOrderHandler)
	router.HandleFunc("/orders/find/{order_id}", findOrderHandler)
	router.HandleFunc("/orders/addItem/{order_id}/{item_id}", addItemHandler)
	router.HandleFunc("/orders/removeItem/{order_id}/{item_id}", removeItemHandler)
	router.HandleFunc("/orders/checkout/{order_id}", checkoutHandler)

	//router.HandleFunc("/orders/send/{message}", sendKafkaMessageHandler)

	//sendKafkaMessageHandler("Hello Deepali")

	fmt.Println("here ... !!")
	port := os.Getenv("PORT")
	fmt.Printf("Current port is: %s\n", port)
	if port == "" {
		port = "8080"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting order service at %s\n", addr)

	fmt.Println("going inside producer")
	producerHandler()
	log.Println("Reached here as wellll")

	log.Fatal(http.ListenAndServe(addr, router))
}

func getOrder(orderID string) (error, *Order) {
	convertDocIDErr, documentID := ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		return convertDocIDErr, nil
	}

	filter := bson.M{"_id": documentID}
	var order Order
	findDocErr := ordersCollection.FindOne(context.Background(), filter).Decode(&order)
	if findDocErr != nil {
		return findDocErr, nil
	}
	order.OrderID = orderID
	return nil, &order
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello welcome to the order service!")
}

// TODO: set to POST method
func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	/*
		Handler for the route responsible for creating a new order given a certain user ID.
	*/
	vars := mux.Vars(r)
	userID := vars["user_id"]
	//fmt.Printf("Creating order for user %s\n", userID)

	order := Order{
		Paid:      false,
		Items:     []string{},
		UserID:    userID,
		TotalCost: 0.0,
	}
	insertResult, mongoInsertErr := ordersCollection.InsertOne(context.Background(), order)
	if mongoInsertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	orderID := insertResult.InsertedID.(primitive.ObjectID).Hex()
	//log.Printf("Inserted document with ID: %v\n", insertResult.InsertedID)
	order.OrderID = orderID

	w.Header().Set("Content-Type", "application/json")
	jsonEncodeErr := json.NewEncoder(w).Encode(order)
	if jsonEncodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	convertDocIDErr, documentID := ConvertStringToMongoID(orderID)
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
	//log.Printf("Deleted %d document(s)\n", result.DeletedCount)
	//fmt.Printf("Deleted order with ID: %s\n", orderID)
}

// TODO: set to GET method
func findOrderHandler(w http.ResponseWriter, r *http.Request) {
	/*
		Handler for the route responsible for finding and returning an order.
	*/
	vars := mux.Vars(r)
	orderID := vars["order_id"]

	findOrderErr, order := getOrder(orderID)
	if findOrderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonEncodeErr := json.NewEncoder(w).Encode(order)
	if jsonEncodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// TODO: set to POST method
func addItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]

	getStockResponse, getStockErr := http.Get(fmt.Sprintf("http://localhost:8082/stock/find/%s", itemID))
	if getStockErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer getStockResponse.Body.Close()

	var item Item
	jsonDecodeErr := json.NewDecoder(getStockResponse.Body).Decode(&item)
	if jsonDecodeErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//if item.Stock == 0 || item.Price == 0.0 {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	convertDocIDErr, documentID := ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderFilter := bson.M{"_id": documentID}
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

// TODO: set to DELETE method
func removeItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_id"]
	itemID := vars["item_id"]

	getStockResponse, getStockErr := http.Get(fmt.Sprintf("http://localhost:8082/stock/find/%s", itemID))
	if getStockErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer getStockResponse.Body.Close()

	var item Item
	jsonDecodeErr := json.NewDecoder(getStockResponse.Body).Decode(&item)
	if jsonDecodeErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	convertDocIDErr, documentID := ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderFilter := bson.M{"_id": documentID}
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

// PAYMENT_MAKE_USERID_ORDERID_TOTALCOST
// PAYMENT_CANCEL_USERID_ORDERID_TOTALCOST

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
	getOrderErr, order := getOrder(orderID)
	if getOrderErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Step 1: Make the payment
	// Compensation: cancel payment
	//paymentSuccess := makePayment(*order)
	//if !paymentSuccess {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}
	//
	//// Step 2: Subtract the stock
	//// Compensation: re-add stock
	//for _, item := range order.Items {
	//	//fmt.Printf("Item at index %d: %+v\n", i, item)
	//	subtractStockSuccess := subtractStock(item, 1)
	//	if !subtractStockSuccess {
	//		w.WriteHeader(http.StatusBadRequest)
	//		return
	//	}
	//}

	// Step 3: Update the order status
	// Compensation: set order to unpaid

	fmt.Print(order.TotalCost)

	//TODO: send kafka message to order saga that local txn is started
	convertDocIDErr, orderDocumentID := ConvertStringToMongoID(orderID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status, updateOrderError := update_order(orderDocumentID)
	if status == Failure {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Print(updateOrderError)
	} else {
		//TODO: send kafka message that local transaction is successful
	}

	status1, updateOrderError1 := on_ack(orderDocumentID)
	if status1 == Failure {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Print(updateOrderError1)
	} else {
		//TODO: send kafka message that local transaction is successful
	}
}

func update_order(orderDocumentID *primitive.ObjectID) (Status, error) {

	status := Status(Pending)
	orderFilter := bson.M{"_id": orderDocumentID}
	orderUpdate := bson.M{
		"$set": bson.M{
			"status": status,
			"paid":   true,
		},
	}
	_, updateOrderErr := ordersCollection.UpdateOne(context.Background(), orderFilter, orderUpdate)
	return status, updateOrderErr
}

func on_ack(orderDocumentID *primitive.ObjectID) (Status, error) {

	status := Status(Completed)
	orderFilter := bson.M{"_id": orderDocumentID}
	orderUpdate := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}
	_, updateOrderErr := ordersCollection.UpdateOne(context.Background(), orderFilter, orderUpdate)
	if updateOrderErr != nil {
		return Failure, updateOrderErr
	}

	return Completed, nil
}

//rollback
func on_nack(orderDocumentID *primitive.ObjectID) (Status, error) {
	status := Status(Rollback_Completed)

	orderFilter := bson.M{"_id": orderDocumentID}
	orderUpdate := bson.M{
		"$set": bson.M{
			"status": status,
			"paid":   false,
		},
	}
	_, updateOrderErr := ordersCollection.UpdateOne(context.Background(), orderFilter, orderUpdate)

	if updateOrderErr != nil {
		return Rollback_Failure, updateOrderErr
	}

	return status, updateOrderErr
}

//func sendKafkaMessageHandler(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	message := vars["message"]
//
//	fmt.Printf("Message to send over Kafka: %s\n", message)
//
//	// to produce messages
//	topic := "stock-syn"
//	partition := 0
//
//	conn, err := kafka.DialLeader(context.Background(), "tcp", "kafka-service:9092", topic, partition)
//	if err != nil {
//		log.Fatal("failed to dial leader:", err)
//	}
//
//	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
//	_, err = conn.Write([]byte(message))
//	if err != nil {
//		log.Fatal("failed to write messages:", err)
//	}
//
//	if err := conn.Close(); err != nil {
//		log.Fatal("failed to close writer:", err)
//	}
//}

//func producerHandler(kafkaWriter *kafka.Writer, msg string) {
//	log.Print("Starting write")
//	msgs := kafka.Message{
//		Key:   []byte(fmt.Sprintf("address-%s", "test")),
//		Value: []byte(msg),
//	}
//	log.Print(msgs)
//	err := kafkaWriter.WriteMessages(context.Background(), msgs)
//
//	log.Print("Reached here")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	log.Print("Reached here as well")
//}
//
//func getKafkaWriter(kafkaURL, topic string) *kafka.Writer {
//	return &kafka.Writer{
//		Addr:     kafka.TCP(kafkaURL),
//		Topic:    topic,
//		Balancer: &kafka.LeastBytes{},
//	}
//}

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}


func producerHandler() {
	// get kafka writer using environment variables.

	fmt.Println("In producerHandler")
	kafkaURL := "kafka-service.kafka:9092"
	topic := "stock-syn"
	partition := 1

	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaURL, topic, partition)
	defer conn.Close()
	if err != nil {
		log.Fatal("failed to dial leader:", err)
		}
	fmt.Println("start producing!!!")

	for i := 0; ; i++ {
		key := fmt.Sprintf("Key-%d", i)
		msg := kafka.Message{Value: []byte("test write")}

		_, err := conn.WriteMessages(msg)
		if err != nil {
			fmt.Print("if err: ")
			fmt.Println(err)
		} else {
			fmt.Println("produced", key)
		}
		time.Sleep(1 * time.Second)
	}
}

//func producerHandler() {
//	// get kafka writer using environment variables.
//
//	kafkaURL := "kafka-service:9092"
//	topic := "stock-syn"
//	writer := newKafkaWriter(kafkaURL, topic)
//	defer writer.Close()
//	fmt.Println("start produccccing ... !!")
//
//	for i := 0; ; i++ {
//		key := fmt.Sprintf("Key-%d", i)
//		msg := kafka.Message{
//			Key:   []byte(key),
//			Value: []byte(fmt.Sprint(uuid.New())),
//		}
//
//		err := writer.WriteMessages(context.Background(), msg)
//		if err != nil {
//			fmt.Print("if err: ")
//			fmt.Println(err)
//		} else {
//			fmt.Println("produced", key)
//		}
//		time.Sleep(1 * time.Second)
//	}
//}
