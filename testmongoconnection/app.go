package main

import (
    "context"
    "fmt"
    "log"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    fmt.Println("Testing mongo connection")

    // Set client options
    clientOptions := options.Client().ApplyURI("mongodb://mongodb-service:27017")

    // Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected")

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
	fmt.Println("error")
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")
}

