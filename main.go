package main

import (
	"context"
	"fmt"
	"log"
	"os"

	fetchdata "github.com/Programmerdin/FinancialDataSite_Go/fetchDataFolder"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
	// var META_CIK string = "0001326801"
	// var AAPL_CIK string = "0000320193"
	var SMRT_CIK string = "0001837014"

	// test, err := fetchdata.GetSubmissionFilesOfCIK(SMRT_CIK)
	// fmt.Println(test)

	// test2, err := fetchdata.ReadJsonFile(test[0])
	// // fmt.Println(test2)

	// test3 := gjson.Get(test2, "filings.recent.accessionNumber")
	// // fmt.Println(test3)
	// fmt.Println(reflect.TypeOf(test3))

	// test4 := gjson.Parse(test2)
	// // fmt.Println(test4)
	// fmt.Println(reflect.TypeOf(test4))

	// var CIK_submission_filePath string = "SEC-files/submissions/CIK" + SMRT_CIK + ".json"
	// test5, err := fetchdata.Parse10K10QmetadataFromSubmissionJsonFile(CIK_submission_filePath)
	// fmt.Println(test5, err)

	// test6, err := fetchdata.Get10K10QMetadataFromSubmissionFilesCIK(SMRT_CIK)
	// fmt.Println(test6)

	fetchdata.Store10K10QmetadataFromSubmissionFilesCIKtoMongoDB(client, SMRT_CIK)
}
