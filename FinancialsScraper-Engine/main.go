package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"

	geteverythinggivencik "github.com/Programmerdin/FinancialDataSite_Go/getEverythingGivenCIK"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to MongoDB
	mongoID := os.Getenv("mongodb_id")
	mongoPassword := os.Getenv("mongodb_password")
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	// Construct the MongoDB URI using the provided credentials
	mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@financialdatasitecluste.scp0c5v.mongodb.net/?retryWrites=true&w=majority", mongoID, mongoPassword)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	// var KO_CIK string = "0000021344"
	var META_CIK string = "0001326801"
	// var AAPL_CIK string = "0000320193"
	// var SMRT_CIK string = "0001837014"
	// var SMRT_accessionNumber1 string = "0001104659-21-068286"

	// var KO_accessionNumber1 string = "0001104659-21-068286"

	// var Meta_BS_filepath string = "SEC-files\\filingSummaryAndRfiles\\0001326801\\0001326801-13-000003\\R2.csv"
	// var Meta_IS_filepath string = "SEC-files\\filingSummaryAndRfiles\\0001326801\\0001326801-13-000003\\R4.csv"

	// combinecsvfiles.GetFinancialStatementsCsvRfilePathsGivenCIK(SMRT_CIK, client)
	geteverythinggivencik.GetEverythingGivenCIK(META_CIK, client)
	// combinecsvfiles.CommonFieldFinderForFinancialStatementRfile(Meta_IS_filepath)

	// combinecsvfiles.GetCsvRfilesIntoArrayVariables(META_CIK, client)

	// // Test OpenAI Connection
	// if err := combinecsvfiles.TestOpenAIConnection(); err != nil {
	// 	log.Fatal("Failed to connect to OpenAI:", err)
	// }
}
