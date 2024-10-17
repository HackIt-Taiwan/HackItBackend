package database

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect() (*mongo.Database, error) {
	clientOptions := options.Client()
	clientOptions.ApplyURI(os.Getenv("MONGO_HOST"))
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	var ctx = context.Background()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")

	return client.Database(os.Getenv("MONGO_DB_NAME")), nil
}
