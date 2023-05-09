package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

func ConvertStringToMongoID(key string) primitive.ObjectID {
	documentID, hexErr := primitive.ObjectIDFromHex(key)
	if hexErr != nil {
		log.Fatal(hexErr)
	}
	return documentID
}

func FindSingleDocument(coll *mongo.Collection, filter interface{}, result interface{}) interface{} {
	err := coll.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No document found with the given filter")
		} else {
			log.Fatal(err)
		}
		return nil
	}
	fmt.Printf("Found document: %+v\n", result)
	return result
}
