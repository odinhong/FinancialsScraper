// package combinecsvfiles

// import (
// 	"context"
// 	"encoding/csv"
// 	"fmt"
// 	"io"
// 	"os"

// 	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// func AddReportDateAndFormToCsvRfile(CsvRfile_FilePath string, client *mongo.Client) error {
// 	// First read just the first row of the CSV file
// 	existingFile, err := os.Open(CsvRfile_FilePath)
// 	if err != nil {
// 		return fmt.Errorf("error opening CSV file for reading: %v", err)
// 	}
// 	defer existingFile.Close()

// 	// Read only the first row to get column count
// 	reader := csv.NewReader(existingFile)
// 	firstRow, err := reader.Read()
// 	if err != nil {
// 		return fmt.Errorf("error reading CSV first row: %v", err)
// 	}

// 	// Get the number of columns from the first row
// 	columnCount := len(firstRow)

// 	//get accession number from filePath
// 	// Get report date and form type from MongoDB
// 	ResultReportDate, ResultFormType, err := FindReportDateAndFormGivenAccessionNumber()

// 	// Create temporary file
// 	tempFile := CsvRfile_FilePath + ".tmp"
// 	outFile, err := os.Create(tempFile)
// 	if err != nil {
// 		return fmt.Errorf("error creating temporary file: %v", err)
// 	}
// 	defer outFile.Close()

// 	writer := csv.NewWriter(outFile)
// 	defer writer.Flush()

// 	// Write report date row
// 	reportDateRow := make([]string, columnCount)
// 	reportDateRow[0] = "Report Date"
// 	for i := 1; i < columnCount; i++ {
// 		reportDateRow[i] = result.ReportDate
// 	}
// 	if err := writer.Write(reportDateRow); err != nil {
// 		return fmt.Errorf("error writing report date row: %v", err)
// 	}

// 	// Write form type row
// 	formTypeRow := make([]string, columnCount)
// 	formTypeRow[0] = "Form Type"
// 	for i := 1; i < columnCount; i++ {
// 		formTypeRow[i] = result.FormType
// 	}
// 	if err := writer.Write(formTypeRow); err != nil {
// 		return fmt.Errorf("error writing form type row: %v", err)
// 	}

// 	// Reset the original file to beginning and copy rest of the content
// 	existingFile.Seek(0, 0)
// 	if _, err := io.Copy(outFile, existingFile); err != nil {
// 		return fmt.Errorf("error copying original content: %v", err)
// 	}

// 	// Replace the original file with the temporary file
// 	if err := os.Rename(tempFile, CsvRfile_FilePath); err != nil {
// 		return fmt.Errorf("error replacing original file: %v", err)
// 	}

// 	return nil
// }

// // FindReportDateAndFormGivenAccessionNumber finds the report date and form type for a given accession number from MongoDB
// func FindReportDateAndFormGivenAccessionNumber(accessionNumber string, client *mongo.Client) (ReportDate string, Form string, err error) {
// 	collection := utilityfunctions.GetMongoDBCollection(client)

// 	projection := bson.D{
// 		primitive.E{Key: "reportdate", Value: 1},
// 		primitive.E{Key: "form", Value: 1},
// 	}

// 	// Find one document with the matching accessionNumber
// 	var result bson.M
// 	err = collection.FindOne(
// 		context.Background(),
// 		bson.D{{Key: "accessionnumber", Value: accessionNumber}},
// 		options.FindOne().SetProjection(projection),
// 	).Decode(&result)

// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return "", "", fmt.Errorf("no document found with accessionNumber: %s", accessionNumber)
// 		}
// 		return "", "", err
// 	}

// 	reportDate, _ := result["reportdate"].(string)
// 	form, _ := result["form"].(string)

// 	return reportDate, form, nil
// }