package parserfiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fetchdata "github.com/Programmerdin/FinancialDataSite_Go/fetchDataFolder"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DownloadRfiles(CIK string, client *mongo.Client) error {
	accessionNumbers, RfileNames, err := RetrieveRfileNamesAndAccessionNumbersFromMongoDB(CIK, client)
	if err != nil {
		fmt.Println("Error RetrieveRfileNamesAndAccessionNumbersFromMongoDB function:", err)
		return err
	}
	downloadLinks, filePaths, err := GenerateDownloadLinksAndFilePathsForRfiles(CIK, accessionNumbers, RfileNames)
	if err != nil {
		fmt.Println("Error GenerateDownloadLinksAndFilePathsForRfiles function:", err)
		return err
	}
	err = fetchdata.DownloadManySECFiles(downloadLinks, filePaths)
	if err != nil {
		fmt.Println("Error DownloadManySECFiles function:", err)
		return err
	}

	return nil
}

func GenerateDownloadLinksAndFilePathsForRfiles(CIK string, accessionNumbers []string, RfileNames []string) ([]string, []string, error) {
	var downloadLinks []string
	var filePaths []string

	for i := 0; i < len(accessionNumbers); i++ {
		filePath := filepath.Join("SEC-files", "filingSummaryAndRfiles", CIK, accessionNumbers[i], RfileNames[i])

		// Check if the file exists already
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			downloadLink := "https://www.sec.gov/Archives/edgar/data/" + CIK + "/" + strings.Replace(accessionNumbers[i], "-", "", -1) + "/" + RfileNames[i]
			downloadLinks = append(downloadLinks, downloadLink)
			filePaths = append(filePaths, filePath)
		} else if err != nil {
			// If there's an error other than the file not existing, return the error
			return nil, nil, err
		}
		// No else needed; if the file exists, we simply don't add it to the slices
	}

	return downloadLinks, filePaths, nil
}

func RetrieveRfileNamesAndAccessionNumbersFromMongoDB(CIK string, client *mongo.Client) ([]string, []string, error) {
	db := client.Database("testDatabase")
	collection := db.Collection("testMetaDataOf10K10Q")
	ctx := context.Background()
	filter := bson.M{
		"cik":               CIK,
		"hasFilingSummary":  true,
		"Rfile_BS_fileName": bson.M{"$exists": true},
	}
	projection := bson.M{
		"accessionNumber":    1,
		"Rfile_BS_fileName":  1,
		"Rfile_IS_fileName":  1,
		"Rfile_CIS_fileName": 1,
		"Rfile_CF_fileName":  1,
		"_id":                0,
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var accessionNumbers []string
	var RfileNames []string

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, nil, err
		}
		// Collect file names and accessionNumbers if they exist in the document
		if fileName, ok := doc["Rfile_BS_fileName"].(string); ok {
			RfileNames = append(RfileNames, fileName)
			accessionNumbers = append(accessionNumbers, doc["accessionNumber"].(string))
		}
		if fileName, ok := doc["Rfile_IS_fileName"].(string); ok {
			RfileNames = append(RfileNames, fileName)
			accessionNumbers = append(accessionNumbers, doc["accessionNumber"].(string))
		}
		if fileName, ok := doc["Rfile_CIS_fileName"].(string); ok {
			RfileNames = append(RfileNames, fileName)
			accessionNumbers = append(accessionNumbers, doc["accessionNumber"].(string))
		}
		if fileName, ok := doc["Rfile_CF_fileName"].(string); ok {
			RfileNames = append(RfileNames, fileName)
			accessionNumbers = append(accessionNumbers, doc["accessionNumber"].(string))
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	return accessionNumbers, RfileNames, nil
}
