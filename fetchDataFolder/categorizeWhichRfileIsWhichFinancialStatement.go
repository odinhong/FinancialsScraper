package fetchdata

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// func GenerateLinksToDownloadFilingSummaryFiles(CIK_slice []string, accessionNumber_slice []string) {
// 	var filingSummaryUrls []string
// 	var base string = SEC-files/companyFilingFiles
// }

func RetrieveCIKAndAccessionNumberThatHaveFilingSummary(CIK string, client *mongo.Client) ([]string, error) {
	// go to mongodb and get two slice of cik and accessionNumber from those docs that have hasFilingSummary = true
	var CIK_slice []string
	var accessionNumber_slice []string

	databaseName := "testDatabase"
	collectionName := "testMetaDataOf10K10Q"
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
			AccessionNumber string `bson:"accessionNumber"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		accessionNumber_slice = append(accessionNumber_slice, result.AccessionNumber)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	fmt.Println(CIK_slice, accessionNumber_slice)
	return accessionNumber_slice, nil
}
