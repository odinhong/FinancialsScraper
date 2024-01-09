package categorizefinancialstatements

import (
	"context"
	"fmt"
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ParseManyFilingSummaryXmlFilesGivenCIK(CIK string, client *mongo.Client) {
	accessionNumbers_slice, err := RetrieveAccessionNumbersThatHaveFilingSummaries(CIK, client)
	if err != nil {
		fmt.Println("Error retrieving accession numbers:", err)
		return
	}

	var filePathsToFilingSummaryFiles []string
	//generate local file paths to FilingSummary.xml files
	for _, accessionNumber := range accessionNumbers_slice {
		filePath := filepath.Join("SEC-files", "filingSummaryAndRfiles", CIK, accessionNumber, "FilingSummary.xml")
		filePathsToFilingSummaryFiles = append(filePathsToFilingSummaryFiles, filePath)
	}

}

func RetrieveAccessionNumbersThatHaveFilingSummaries(CIK string, client *mongo.Client) ([]string, error) {
	databaseName := "testDatabase"
	collectionName := "testMetaDataOf10K10Q"
	collection := client.Database(databaseName).Collection(collectionName)

	filter := bson.M{"hasFilingSummary": true, "cik": CIK}
	projection := bson.M{"accessionNumber": 1, "_id": 0} // Project only the accessionNumber

	// Use a context with timeout or a background context as needed
	ctx := context.Background()

	// Perform the query to find all matching documents
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accessionNumbers []string // Slice to hold the accession numbers

	// Iterate through the cursor and collect accession numbers and cik values
	for cursor.Next(ctx) {
		var result struct {
			AccessionNumber string `bson:"accessionNumber"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		accessionNumbers = append(accessionNumbers, result.AccessionNumber)
	}

	// Check if the cursor encountered any errors during iteration
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return accessionNumbers, nil
}
