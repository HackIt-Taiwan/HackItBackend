package database

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Db *mongo.Database

func Connect() (*mongo.Database, error) {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_HOST"))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")

	Db = client.Database(os.Getenv("MONGO_DB_NAME"))
	return Db, nil
}
