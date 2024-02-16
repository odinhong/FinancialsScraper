package combinecsvfiles

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RetrieveFinancialStatementMetaDataDocsOldestToNewestReportingDate(CIK string, client *mongo.Client) {
	databaseName := os.Getenv("DATABASE_NAME")
	collectionName := os.Getenv("COLLECTION_NAME")
	collection := client.Database(databaseName).Collection(collectionName)

	// Finding multiple documents with the specified CIK and Sorting by reportdate in ascending order and filtering by CIK
	cur, err := collection.Find(context.Background(), bson.D{{"cik", CIK}}, options.Find().SetSort(bson.D{{"reportdate", 1}}))
	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(context.Background())

	// Iterate through the cursor
	var results []bson.M
	if err = cur.All(context.Background(), &results); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Documents for CIK %s sorted from oldest to newest: %+v\n", CIK, results)
}
