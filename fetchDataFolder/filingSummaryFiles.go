package fetchdata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// func testing123(accessionNumbers []string, CIKs []string) {
// 	var base_url = "https://www.sec.gov/Archives/edgar/data/"
// 	var complete_url = base_url + CIKs[0] + "/" + accessionNumbers[0] + "/index.json"
// 	//get the response body from complete_url

// }

func CheckOneFilingIndexJsonForExistenceOfFilingSummary(CIK string, accessionNumber string, client *mongo.Client) {
	userAgent := os.Getenv("USER_AGENT")
	companyName := os.Getenv("COMPANY_NAME")
	email := os.Getenv("EMAIL")

	var SEC_indexJson_url = "https://www.sec.gov/Archives/edgar/data/" + CIK + "/" + strings.Replace(accessionNumber, "-", "", -1) + "/index.json"

	// Create a new request
	req, err := http.NewRequest("GET", SEC_indexJson_url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", fmt.Sprintf("%s - %s (mailto:%s)", userAgent, companyName, email))

	// Create a new HTTP client and send the request
	HTTPclient := &http.Client{}
	resp, err := HTTPclient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonString := string(body)
	items := gjson.Get(jsonString, "directory.item")
	hasFilingSummary := false

	items.ForEach(func(key, value gjson.Result) bool {
		name := value.Get("name").String()
		if strings.Contains(strings.ToLower(name), "filingsummary") {
			hasFilingSummary = true
			return false // Stop iterating after finding FilingSummary
		}
		return true // Continue iterating
	})

	databaseName := "testDatabase"
	collectionName := "testMetaDataOf10K10Q"
	collection := client.Database(databaseName).Collection(collectionName)
	ctx := context.Background()
	filter := bson.M{"accessionNumber": accessionNumber, "cik": CIK}

	// Output the result and update the 'hasFilingSummary' field to Mongo
	if hasFilingSummary {
		update := bson.M{"$set": bson.M{"hasFilingSummary": true}}
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		update := bson.M{"$set": bson.M{"hasFilingSummary": false}}
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func GetListOfFilingsThatHaveNotCheckedExistenceOfFilingSummary(client *mongo.Client) ([]string, []string, error) {
	databaseName := "testDatabase"
	collectionName := "testMetaDataOf10K10Q"
	collection := client.Database(databaseName).Collection(collectionName)

	// Create a filter to find documents where 'hasFilingSummary' does not exist
	filter := bson.M{"hasFilingSummary": bson.M{"$exists": false}}
	projection := bson.M{"accessionNumber": 1, "cik": 1, "_id": 0} // Project only the accessionNumber and cik fields

	// Use a context with timeout or a background context as needed
	ctx := context.Background()

	// Perform the query to find all matching documents
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	var accessionNumbers []string // Slice to hold the accession numbers
	var ciks []string             // Slice to hold the cik values

	// Iterate through the cursor and collect accession numbers and cik values
	for cursor.Next(ctx) {
		var result struct {
			AccessionNumber string `bson:"accessionNumber"`
			CIK             string `bson:"cik"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, nil, err
		}
		accessionNumbers = append(accessionNumbers, result.AccessionNumber)
		ciks = append(ciks, result.CIK)
	}

	// Check if the cursor encountered any errors during iteration
	if err := cursor.Err(); err != nil {
		return nil, nil, err
	}

	return accessionNumbers, ciks, nil
}
