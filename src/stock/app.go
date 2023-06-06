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

type ItemChange struct {
	itemID *uuid.UUID
	amount int64
}

var numInstances *int64
var clients []*mongo.Client
var collections []*mongo.Collection

func main() {
	var instanceNumErr error
	instanceNumErr, numInstances = shared.GetNumOfServices(shared.StockService)
	if instanceNumErr != nil {
		log.Fatal(instanceNumErr)
	}

	go shared.SetUpKafkaListener(
		[]string{"stock"}, false,
		func(message *shared.SagaMessage) (*shared.SagaMessage, string) {

			returnMessage := shared.SagaMessageConvertStartToEnd(message)

			// TODO: remove code duplication

			if message.Name == "START-SUBTRACT-STOCK" {
				changes := make([]ItemChange, len(message.Order.Items))

				for i, stringID := range message.Order.Items {
					// ignore error, will not happen
					_, itemID := shared.ConvertStringToUUID(stringID)

					changes[i] = ItemChange{
						itemID: itemID,
						amount: 1,
					}
				}

				clientError, serverError := subtract(changes)
				if clientError != nil || serverError != nil {
					returnMessage.Name = "ABORT-CHECKOUT-SAGA"
				}

				return returnMessage, "stock-ack"
			}

			if message.Name == "START-READD-STOCK" {
				changes := make([]ItemChange, len(message.Order.Items))

				for i, stringID := range message.Order.Items {
					// ignore error, will not happen
					_, itemID := shared.ConvertStringToUUID(stringID)

					changes[i] = ItemChange{
						itemID: itemID,
						amount: 1,
					}
				}

				clientError, serverError := add(changes)
				if clientError != nil || serverError != nil {
					returnMessage.Name = "ABORT-CHECKOUT-SAGA"
				}

				return returnMessage, "stock-ack"
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
	for i := 0; i < int(*numInstances); i++ {
		defer clients[i].Disconnect(ctx)
	}

	router := mux.NewRouter()
	router.HandleFunc("/find/{item_id}", findHandler)
	router.HandleFunc("/subtract/{item_id}/{amount}", subtractHandler)
	router.HandleFunc("/add/{item_id}/{amount}", addHandler)
	router.HandleFunc("/item/create/{price}", createHandler)
	router.HandleFunc("/", defaultHandler)

	port := os.Getenv("PORT")
	fmt.Printf("Current port is: %s\n", port)
	if port == "" {
		port = "8082"
	}

	// Set the listening address and port for the server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting stock service at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func setupDBConnections(ctx context.Context) error {
	clients := make([]*mongo.Client, *numInstances)
	collections := make([]*mongo.Collection, *numInstances)

	for i := 0; i < int(*numInstances); i++ {
		mongoURL := fmt.Sprintf("mongodb://stockdb-service-%d:27017", i)
		fmt.Printf("%d MongoDB URL: %s", i, mongoURL)
		var err error
		var client *mongo.Client
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))

		if err != nil {
			return err
		}
		clients[i] = client
		collections[i] = client.Database("stock").Collection("stock")
	}
	return nil
}

func getItem(documentID *uuid.UUID) (error, *shared.Item) {
	stockCollection := getStockCollection(documentID)

	var item shared.Item
	err := stockCollection.FindOne(context.Background(), bson.M{"_id": documentID}).Decode(&item)
	if err != nil {
		return err, nil
	}
	item.ID = *documentID
	item.ItemID = documentID.String()
	return nil, &item
}

func getStockCollection(itemID *uuid.UUID) *mongo.Collection {
	databaseNum := shared.HashUUID(*itemID)
	return collections[databaseNum]
}

// Functions only used by http

func findHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]
	// fmt.Printf("item ID: %s", itemID)

	convertDocIDErr, documentID := shared.ConvertStringToUUID(itemID)
	if convertDocIDErr != nil {
		fmt.Println("CONVERT DOC ERROR")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fmt.Printf("Find: %s\n", itemID)
	findErr, item := getItem(documentID)
	if findErr != nil {
		fmt.Println("GET ITEM ERROR", findErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	jsonEncodeErr := json.NewEncoder(w).Encode(item)
	if jsonEncodeErr != nil {
		fmt.Println("JSON ENCODE ERROR")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("defualt greeter of stock")
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	price := vars["price"]
	err, PriceInt := shared.ConvertStringToInt(price)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// fmt.Printf("Creating item with price %s\n", price)
	documentID := shared.GetNewID()
	stock := shared.Item{
		ID:     documentID,
		ItemID: documentID.String(),
		Stock:  0,
		Price:  *PriceInt,
	}

	stockCollection := getStockCollection(&documentID)
	_, insertErr := stockCollection.InsertOne(context.Background(), stock)
	if insertErr != nil {
		fmt.Println(insertErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonEncodeErr := json.NewEncoder(w).Encode(stock)
	if jsonEncodeErr != nil {
		fmt.Println(jsonEncodeErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// Functions used by http and kafka

func subtractHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]
	amount := vars["amount"]
	convertIntErr, intAmount := shared.ConvertStringToInt(amount)
	if convertIntErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	convertDocIDErr, documentID := shared.ConvertStringToUUID(itemID)
	if convertDocIDErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientError, serverError := subtract([]ItemChange{{
		itemID: documentID,
		amount: *intAmount,
	}})

	if clientError != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if serverError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func subtract(changes []ItemChange) (clientError error, serverError error) {
	amountDone := 0
	changesDone := make([]ItemChange, len(changes))

	for i, change := range changes {
		getItemErr, item := getItem(change.itemID)
		if getItemErr != nil {
			clientError = getItemErr
			break
		}
		if item.Stock < change.amount {
			clientError = errors.New("not enough stock to subtract")
			break
		}
		update := bson.M{
			"$inc": bson.M{
				"stock": -change.amount,
			},
		}

		stockCollection := getStockCollection(change.itemID)
		result := shared.UpdateRecord(stockCollection, bson.M{"_id": change.itemID}, update)
		if result.Err() != nil {
			log.Printf("Update stock error: %s", result.Err())
			serverError = result.Err()
			break
		}

		amountDone++
		changesDone[i] = change
	}

	for i := 0; i < amountDone; i++ {
		changeDone := changesDone[i]

		update := bson.M{
			"$inc": bson.M{
				"stock": changeDone.amount,
			},
		}

		stockCollection := getStockCollection(changeDone.itemID)
		result := shared.UpdateRecord(stockCollection, bson.M{"_id": changeDone.itemID}, update)
		if result.Err() != nil {
			log.Printf("Redo stock error: %s", result.Err())
			serverError = result.Err()
		}
	}

	return
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["item_id"]
	amount := vars["amount"]
	convIntErr, intAmount := shared.ConvertStringToInt(amount)
	if convIntErr != nil {
		log.Print(convIntErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	convStringErr, documentID := shared.ConvertStringToUUID(itemID)
	if convStringErr != nil {
		log.Print(convStringErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	clientError, serverError := add([]ItemChange{{
		itemID: documentID,
		amount: *intAmount,
	}})

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

func add(changes []ItemChange) (clientError error, serverError error) {
	for _, change := range changes {
		filter := bson.M{"_id": change.itemID}
		update := bson.M{
			"$inc": bson.M{
				"stock": change.amount,
			},
		}
		stockCollection := getStockCollection(change.itemID)
		result := shared.UpdateRecord(stockCollection, filter, update)
		if result.Err() != nil {
			log.Print(result.Err())
			serverError = result.Err()
			return
		}
	}
	return
}
