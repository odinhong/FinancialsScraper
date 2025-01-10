package main

import (
	"context"
	"fmt"
	"log"
	"os"

	api "financialscraper/API"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func initMongoDB() error {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	// Get MongoDB credentials from env
	mongoID := os.Getenv("mongodb_id")
	mongoPassword := os.Getenv("mongodb_password")

	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	// Construct the MongoDB URI using the provided credentials
	mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@financialdatasitecluste.scp0c5v.mongodb.net/?retryWrites=true&w=majority",
		mongoID,
		mongoPassword)

	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	var err2 error
	mongoClient, err2 = mongo.Connect(context.TODO(), opts)
	if err2 != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err2)
	}

	// Send a ping to confirm a successful connection
	if err := mongoClient.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return nil
}

func main() {
	// Initialize MongoDB connection
	if err := initMongoDB(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.TODO()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v\n", err)
		}
	}()

	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	api.SetupRoutes(r, mongoClient)

	// Run server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Server starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
