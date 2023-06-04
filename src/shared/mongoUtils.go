package shared

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateRecord(sessCtx context.Context, collection *mongo.Collection, filter interface{}, update interface{}) *mongo.SingleResult {
	options := options.FindOneAndUpdate().SetUpsert(true)
	result := collection.FindOneAndUpdate(context.Background(), filter, update, options)
	return result
}
