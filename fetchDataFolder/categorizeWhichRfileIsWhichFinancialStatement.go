package fetchdata

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RetrieveCIKAndAccessionNumberThatHaveFilingSummary(client *mongo.Client) ([]string, []string, error) {
	// go to mongodb and get two slice of cik and accessionNumber from those docs that have hasFilingSummary = true
	var CIK_slice []string
	var accessionNumber_slice []string

	databaseName := "testDatabase"
	collectionName := "testMetaDataOf10K10Q"
	collection := client.Database(databaseName).Collection(collectionName)
	filter := bson.M{"hasFilingSummary": true}

	ctx := context.TODO()
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			CIK             string `bson:"cik"`
			AccessionNumber string `bson:"accessionNumber"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, nil, err
		}
		CIK_slice = append(CIK_slice, result.CIK)
		accessionNumber_slice = append(accessionNumber_slice, result.AccessionNumber)
	}

	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	fmt.Println(CIK_slice, accessionNumber_slice)
	return CIK_slice, accessionNumber_slice, nil
}
