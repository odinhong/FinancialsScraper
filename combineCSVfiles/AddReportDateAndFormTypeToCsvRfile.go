package combinecsvfiles

import (
	"context"
	"fmt"
	"os"

	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddReportDateAndFormToCsvRfile(CsvRfile_FilePath string, client *mongo.Client) error {
	CsvRfile, err := os.OpenFile(CsvRfile_FilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening CSV Rfile: %v", err)
	}
	defer CsvRfile.Close()

	return nil
}

// FindReportDateAndFormGivenAccessionNumber finds the report date and form type for a given accession number from MongoDB
func FindReportDateAndFormGivenAccessionNumber(accessionNumber string, client *mongo.Client) (ReportDate string, Form string, err error) {
	collection := utilityfunctions.GetMongoDBCollection(client)

	projection := bson.D{
		primitive.E{Key: "reportdate", Value: 1},
		primitive.E{Key: "form", Value: 1},
	}

	// Find one document with the matching accessionNumber
	var result bson.M
	err = collection.FindOne(
		context.Background(),
		bson.D{{Key: "accessionnumber", Value: accessionNumber}},
		options.FindOne().SetProjection(projection),
	).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", "", fmt.Errorf("no document found with accessionNumber: %s", accessionNumber)
		}
		return "", "", err
	}

	// Extract the values
	reportDate, _ := result["reportdate"].(string)
	form, _ := result["form"].(string)

	return reportDate, form, nil
}
