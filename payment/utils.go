package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ConvertStringToMongoID(key string) (error, *primitive.ObjectID) {
	documentID, hexErr := primitive.ObjectIDFromHex(key)
	if hexErr != nil {
		return hexErr, nil
	}
	return nil, &documentID
}

func ConvertStringToFloat(number string) (error, *float64) {
	float, convErr := strconv.ParseFloat(number, 64)
	if convErr != nil {
		return convErr, nil
	}
	return nil, &float
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
