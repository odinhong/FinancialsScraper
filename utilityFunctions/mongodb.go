package utilityfunctions

import (
	"os"

	"go.mongodb.org/mongo-driver/mongo"
)

// GetCollection returns a MongoDB collection using environment variables
func GetMongoDBCollection(client *mongo.Client) *mongo.Collection {
	databaseName := os.Getenv("DATABASE_NAME")
	collectionName := os.Getenv("COLLECTION_NAME")
	return client.Database(databaseName).Collection(collectionName)
}
