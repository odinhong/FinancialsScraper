package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	geteverythinggivencik "github.com/Programmerdin/FinancialDataSite_Go/getEverythingGivenCIK"
	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
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

	// fetchdata.Store10K10QmetadataFromSubmissionFilesCIKtoMongoDB(KO_CIK, client)

	// fetchdata.CheckAllFilingIndexJsonForExistenceOfFilingSummary(SMRT_CIK, client)

	// fetchdata.RetrieveCIKAndAccessionNumberThatHaveFilingSummary(SMRT_CIK, client)

	// var SMRT_FilingSummary_Link string = "https://www.sec.gov/Archives/edgar/data/0001837014/000110465921105196/FilingSummary.xml"
	// var SMRT_FilingSummary_filePath string = "SEC-files\\filingSummaryAndRfiles\\0001837014\\0001104659-21-105196\\FilingSummary.xml"
	// fetchdata.DownloadOneSECFile(SMRT_FilingSummary_Link, SMRT_FilingSummary_filePath)

	// fetchdata.DownloadFilingSummaryFiles(SMRT_CIK, client)

	// var XMLfilePath string = "SEC-files\\filingSummaryAndRfiles\\0001837014\\0000950170-23-006749\\FilingSummary.xml"
	// categorziefinancialstatements.PasrseFilingSummaryXMLcontent(XMLfilePath)

	// categorizefinancialstatements.ParseManyFilingSummaryXmlFilesAndSaveToMongoGivenCIK(SMRT_CIK, client)

	// parserfiles.DownloadRfiles(SMRT_CIK, client)

	// var Smrt_test_accesseionNumber string = "0000950170-22-004604"
	// var smrt_rfilename string = "R4.htm"
	// parserfiles.ParseRfileAndSaveAsCSV(SMRT_CIK, Smrt_test_accesseionNumber, smrt_rfilename)

	// var KO_test_accessionNumber string = "0001047469-10-004416"
	// var ko_rfilename string = "R1.xml"
	// parserfiles.ParseXmlRfile(KO_CIK, KO_test_accessionNumber, ko_rfilename)

	//SEC-files/filingSummaryAndRfiles/0000021344/0000021344-21-000014

	// combinecsvfiles.GetCSVfilepathsInOrder(SMRT_CIK, client)
	geteverythinggivencik.GetEverythingGivenCIK(META_CIK, client)
	// combinecsvfiles.Tester(SMRT_CIK, client)
	utilityfunctions.ConvertStringDatesIntoIntNumberDates("Sep. 30, 2011")
}
