<<<<<<< HEAD:src/shared/conversion.go
package shared

import (
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

func ConvertStringToInt(number string) (error, *int64) {
	integerNum, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return err, nil
	}
	return nil, &integerNum
}
=======
package main
>>>>>>> main:stock/utils.go
