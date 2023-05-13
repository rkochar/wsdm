package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"strconv"
)

func ConvertStringToMongoID(key string) primitive.ObjectID {
	documentID, hexErr := primitive.ObjectIDFromHex(key)
	if hexErr != nil {
		log.Fatal(hexErr)
	}
	return documentID
}

func ConvertStringToInt(number string) int64 {
	integerNum, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return integerNum
}
