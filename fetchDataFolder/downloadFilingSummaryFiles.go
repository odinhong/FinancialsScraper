package fetchdata

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func DownloadFilingSummaryFiles(CIK string, client *mongo.Client) error {
	accessionNumber_slice, err := RetrieveAccessionNumbersThatHaveFilingSummary(CIK, client)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var accessionNumbersToDownloadFilingSummary []string
	var filePathsOfFilingSummaryFilesToDownload []string
	for _, accessionNumber := range accessionNumber_slice {
		filePath := filepath.Join("SEC-files", "filingSummaryAndRfiles", CIK, accessionNumber, "FilingSummary.xml")
		//check if the file exists locally
		if _, err := os.Stat(filePath); err == nil {
			// File exists
			fmt.Println("File exists:", filePath)
		} else if os.IsNotExist(err) {
			// File does not exist, add it to the list of files to download
			accessionNumbersToDownloadFilingSummary = append(accessionNumbersToDownloadFilingSummary, accessionNumber)
			filePathsOfFilingSummaryFilesToDownload = append(filePathsOfFilingSummaryFilesToDownload, filePath)
		} else {
			// Error occurred while checking file existence
			fmt.Println("Error occurred while checking file existence:", err)
		}
	}

	downloadUrlsOfFilingSummaryFiles := GenerateLinksToDownloadFilingSummaryFiles(CIK, accessionNumbersToDownloadFilingSummary)
	DownloadManySECFiles(downloadUrlsOfFilingSummaryFiles, filePathsOfFilingSummaryFilesToDownload)
	return nil
}

func GenerateLinksToDownloadFilingSummaryFiles(CIK string, accessionNumber_slice []string) []string {
	var filingSummaryUrls []string
	var baseUrl string = "https://www.sec.gov/Archives/edgar/data/" + CIK + "/"

	for _, accessionNumber := range accessionNumber_slice {
		filingSummaryUrls = append(filingSummaryUrls, baseUrl+strings.Replace(accessionNumber, "-", "", -1)+"/FilingSummary.xml")
	}

	return filingSummaryUrls
}

func RetrieveAccessionNumbersThatHaveFilingSummary(CIK string, client *mongo.Client) ([]string, error) {
	// go to mongodb and get slice of accessionNumber from those docs that have hasFilingSummary = true
	var accessionNumber_slice []string

	databaseName := os.Getenv("DATABASE_NAME")
	collectionName := os.Getenv("COLLECTION_NAME")
	// databaseName := "testDatabase"
	// collectionName := "testMetaDataOf10K10Q"
	collection := client.Database(databaseName).Collection(collectionName)
	filter := bson.M{"hasFilingSummary": true, "cik": CIK}

	ctx := context.TODO()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			AccessionNumber string `bson:"accessionnumber"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		accessionNumber_slice = append(accessionNumber_slice, result.AccessionNumber)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return accessionNumber_slice, nil
}
