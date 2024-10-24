package utils

import (
	"context" // 用於傳遞上下文
	"fmt"
	"hackitbackend/platform/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Counter struct {
	Type  string `bson:"type"`
	Value int    `bson:"value"`
}

func GetNextUserID(ctx context.Context, type_ string) (int, error) {
	filter := bson.M{"type": type_}
	update := bson.M{"$inc": bson.M{"value": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	var counter Counter
	err := database.Db.Collection("counters").FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("failed to increment user ID: %v", err)
	}

	return counter.Value, nil
}
