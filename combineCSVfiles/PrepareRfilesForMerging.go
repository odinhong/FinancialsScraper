package combinecsvfiles

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetCsvRfilesIntoArrayVariables(CIK string, client *mongo.Client) (BalanceSheetArrays [][][]string, IncomeStatementArrays [][][]string, ComprehensiveIncomeStatementArrays [][][]string, CashflowStatementArrays [][][]string, err error) {
	MongoDocs, err := RetrieveFinancialStatementMetaDataDocsOldestToNewestReportDate(CIK, client)
	if err != nil {
		fmt.Println("RetrieveFinancialStatementMetaDataDocsOldestToNewestReportDate", err)
		return nil, nil, nil, nil, err
	}
	BSfilePaths, ISfilePaths, CISfilePaths, CFfilePaths, err := GenerateFilePathsOfCSVfilesOfFinancialStatementsGivenMongoDocs(MongoDocs)
	if err != nil {
		fmt.Println("GenerateFilePathsOfCSVfilesOfFinancialStatementsGivenMongoDocs", err)
		return nil, nil, nil, nil, err
	}

	//open csv file and read it into memory
	BS_arrays := [][][]string{}
	for _, filePath := range BSfilePaths {
		if filePath == "" {
			continue
		}
		financialStatement, err := utilityfunctions.ReadCsvFileToArray(filePath)
		if err != nil {
			fmt.Println("Error reading CSV file:", err)
			return nil, nil, nil, nil, err
		}
		BS_arrays = append(BS_arrays, financialStatement)
	}

	IS_arrays := [][][]string{}
	for _, filePath := range ISfilePaths {
		if filePath == "" {
			continue
		}
		financialStatement, err := utilityfunctions.ReadCsvFileToArray(filePath)
		if err != nil {
			fmt.Println("Error reading CSV file:", err)
			return nil, nil, nil, nil, err
		}
		IS_arrays = append(IS_arrays, financialStatement)
	}

	CIS_arrays := [][][]string{}
	for _, filePath := range CISfilePaths {
		if filePath == "" {
			continue
		}
		financialStatement, err := utilityfunctions.ReadCsvFileToArray(filePath)
		if err != nil {
			fmt.Println("Error reading CSV file:", err)
			return nil, nil, nil, nil, err
		}
		CIS_arrays = append(CIS_arrays, financialStatement)
	}

	CF_arrays := [][][]string{}
	for _, filePath := range CFfilePaths {
		if filePath == "" {
			continue
		}
		financialStatement, err := utilityfunctions.ReadCsvFileToArray(filePath)
		if err != nil {
			fmt.Println("Error reading CSV file:", err)
			return nil, nil, nil, nil, err
		}
		CF_arrays = append(CF_arrays, financialStatement)
	}
	fmt.Println("BS_arrays", BS_arrays[0], BS_arrays[1], BS_arrays[2], BS_arrays[3])

	return BS_arrays, IS_arrays, CIS_arrays, CF_arrays, nil
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

	//make sure the slice lengths are the same if they are not 0, Not checking CIS since not all docs have CIS
	if len(BSfilePaths) != len(ISfilePaths) || len(BSfilePaths) != len(CFfilePaths) {
		error = fmt.Errorf("length of BSfilePaths, ISfilePaths, CFfilePaths is not the same")
	}
	return BSfilePaths, ISfilePaths, CISfilePaths, CFfilePaths, error
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
