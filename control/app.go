package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Transaction struct {
	SagaId  string  `json:"saga_id"`
	Service Service `json:"service"`
	Status  string  `json:"status"`
}

type Saga struct {
	SagaId string `json:"saga_id"`
	Status string `json:"status"`
}

var client *mongo.Client
var sagaCollection *mongo.Collection
var transactionCollection *mongo.Collection

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://stockdb-svc-0:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("saga")
	sagaCollection = db.Collection("saga")
	transactionCollection = db.Collection("transaction")

}

func start_saga() (error, *primitive.ObjectID) {

	user := Saga{
		Status: Pending,
	}
	result, insertionError := sagaCollection.InsertOne(context.Background(), user)
	if insertionError != nil {
		return insertionError, nil
	}
	err, sagaId := ConvertStringToMongoID(result.InsertedID.(primitive.ObjectID).Hex())
	return err, sagaId
}

func add_transaction(sagaId primitive.ObjectID, service string) (error, *primitive.ObjectID) {

	txn := Transaction{
		Status:  Pending,
		SagaId:  sagaId.String(),
		Service: Service(service),
	}
	result, insertionError := transactionCollection.InsertOne(context.Background(), txn)
	if insertionError != nil {
		return insertionError, nil
	}
	err, txnId := ConvertStringToMongoID(result.InsertedID.(primitive.ObjectID).Hex())
	return err, txnId
}

func handle_order() {

}
