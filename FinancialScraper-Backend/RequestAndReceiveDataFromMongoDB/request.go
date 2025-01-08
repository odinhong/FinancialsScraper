package requestandreceivedatafrommongodb

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetFinancialStatementsByCIK queries MongoDB for financial statements matching the given CIK
func GetFinancialStatementsByCIK(client *mongo.Client, CIK string) ([]bson.M, error) {
	dbName := "testDatabase2" // Hardcoding the known database name
	collectionName := os.Getenv("combinedFinancialStatementsCollection")
	if collectionName == "" {
		collectionName = "combinedFinancialStatements" // fallback to default
	}

	collection := client.Database(dbName).Collection(collectionName)

	// Create a filter for documents with matching CIK
	filter := bson.D{{"cik", CIK}}

	// Query the collection with the CIK filter
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("MongoDB Find error: %v", err)
	}
	defer cursor.Close(context.Background())

	// Iterate over the results
	var results []bson.M
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return nil, fmt.Errorf("cursor error: %v", err)
	}

	return results, nil
}

// HandleGetFinancialStatements is the HTTP handler for getting financial statements
func HandleGetFinancialStatements(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get CIK from URL parameter
		cik := c.Param("cik")

		// Get financial statements from MongoDB
		results, err := GetFinancialStatementsByCIK(client, cik)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(results) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No financial statements found for this CIK"})
			return
		}

		// Return the results
		c.JSON(http.StatusOK, results)
	}
}
