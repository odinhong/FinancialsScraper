package combinecsvfiles

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetCSVfilepathsInOrder returns the file paths of the Balance Sheet, Income Statement, Cash Flow Statement, and Comprehensive Income Statement CSV files
// in the order of oldest to newest report date given a CIK and a MongoDB client.
func GetCSVfilepathsInOrder(CIK string, client *mongo.Client) (BSfilePaths []string, ISfilePaths []string, CISfilePaths []string, CFfilePaths []string, err error) {
	MongoDocs, err := RetrieveFinancialStatementMetaDataDocsOldestToNewestReportDate(CIK, client)
	if err != nil {
		fmt.Println("RetrieveFinancialStatementMetaDataDocsOldestToNewestReportDate", err)
	}
	BSfilepaths, ISfilepaths, CISfilepaths, CFfilepaths, err := GenerateFilePathsOfCSVfilesOfFinancialStatementsGivenMongoDocs(MongoDocs)
	if err != nil {
		fmt.Println("GenerateFilePathsOfCSVfilesOfFinancialStatementsGivenMongoDocs", err)
	}

	// Print first file path of each type if available
	if len(BSfilepaths) > 0 {
		fmt.Printf("Sample BS file: %s\n", BSfilepaths[0])
	}
	if len(ISfilepaths) > 0 {
		fmt.Printf("Sample IS file: %s\n", ISfilepaths[0])
	}
	if len(CISfilepaths) > 0 {
		fmt.Printf("Sample CIS file: %s\n", CISfilepaths[0])
	}
	if len(CFfilepaths) > 0 {
		fmt.Printf("Sample CF file: %s\n", CFfilepaths[0])
	}

	return BSfilepaths, ISfilepaths, CISfilepaths, CFfilepaths, err
}

func getCsvFilePath(basePath string, CIK string, accessionNumber string, value interface{}) string {
	// Attempt to assert that 'value' is of type string.
	if valueString, ok := value.(string); ok && valueString != "" {
		// If 'ok' is true, the assertion succeeded, and 'valueString' is now a string.
		// Also check if the 'valueString' is not empty.
		csvFilename := strings.TrimSuffix(valueString, ".htm") + ".csv" // Create the new filename.
		return filepath.Join(basePath, CIK, accessionNumber, csvFilename)
	}
	return ""
}

func GenerateFilePathsOfCSVfilesOfFinancialStatementsGivenMongoDocs(MongoDocs []bson.M) (BSfilePaths []string, ISfilePaths []string, CISfilePaths []string, CFfilePaths []string, error error) {
	basefilepath := filepath.Join("SEC-files", "filingSummaryAndRfiles")
	for _, MongoDoc := range MongoDocs {
		accessionNumber, _ := MongoDoc["accessionnumber"].(string) // Assuming accessionnumber is always present and is a string
		CIK, _ := MongoDoc["cik"].(string)                         // Get CIK from MongoDoc

		BSfilePaths = append(BSfilePaths, getCsvFilePath(basefilepath, CIK, accessionNumber, MongoDoc["Rfile_BS_fileName"]))
		ISfilePaths = append(ISfilePaths, getCsvFilePath(basefilepath, CIK, accessionNumber, MongoDoc["Rfile_IS_fileName"]))
		CISfilePaths = append(CISfilePaths, getCsvFilePath(basefilepath, CIK, accessionNumber, MongoDoc["Rfile_CIS_fileName"]))
		CFfilePaths = append(CFfilePaths, getCsvFilePath(basefilepath, CIK, accessionNumber, MongoDoc["Rfile_CF_fileName"]))
	}

	//make sure the slice lengths are the same if they are not 0
	if len(BSfilePaths) != len(ISfilePaths) || len(BSfilePaths) != len(CISfilePaths) || len(BSfilePaths) != len(CFfilePaths) {
		error = fmt.Errorf("length of BSfilePaths, ISfilePaths, CISfilePaths, CFfilePaths is not the same")
	}
	return BSfilePaths, ISfilePaths, CISfilePaths, CFfilePaths, error
}

func RetrieveFinancialStatementMetaDataDocsOldestToNewestReportDate(CIK string, client *mongo.Client) ([]bson.M, error) {
	databaseName := os.Getenv("DATABASE_NAME")
	collectionName := os.Getenv("10K10QMetaDataCollection")
	collection := client.Database(databaseName).Collection(collectionName)

	// Finding multiple documents with the specified CIK and Sorting by reportdate in ascending order(old to new) and filtering by CIK
	cur, err := collection.Find(context.Background(), bson.D{primitive.E{Key: "cik", Value: CIK}}, options.Find().SetSort(bson.D{primitive.E{Key: "reportdate", Value: 1}}))
	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(context.Background())

	// Iterate through the cursor
	var MongoDocs []bson.M
	if err = cur.All(context.Background(), &MongoDocs); err != nil {
		log.Fatal(err)
	}

	return MongoDocs, err
}
