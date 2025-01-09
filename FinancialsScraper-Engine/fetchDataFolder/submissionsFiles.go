package fetchdata

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FilingMetaData struct {
	CIK                string `json:"cik"`
	AccessionNumber    string `json:"accessionnumber"`
	FilingDate         string `json:"filingdate"`
	ReportDate         string `json:"reportdate"`
	AcceptanceDateTime string `json:"acceptancedatetime"`
	Act                string `json:"act"`
	Form               string `json:"form"`
	FileNumber         string `json:"filenumber"`
	FilmNumber         string `json:"filmnumber"`
	Items              string `json:"items"`
	Size               string `json:"size"`
}

func Store10K10QmetadataFromSubmissionFilesCIKtoMongoDB(CIK string, client *mongo.Client) error {
	metadataSlice, err := Get10K10QMetadataFromSubmissionFilesGivenCIK(CIK)
	if err != nil {
		return err
	}
	// fmt.Println("metadataSlice:", metadataSlice)

	collection := utilityfunctions.GetMongoDBCollection(client)

	// Create a context with timeout or a background context as needed
	ctx := context.Background()

	var wg sync.WaitGroup
	errorChannel := make(chan error, len(metadataSlice)) // Buffer error channel to the size of metadataSlice

	// Insert or update each meta data record into MongoDB concurrently
	for _, metaData := range metadataSlice {
		wg.Add(1)                    // Increment the WaitGroup counter
		go func(md FilingMetaData) { // Pass metaData as a local variable to the goroutine
			defer wg.Done() // Decrement the counter when the goroutine completes

			filter := bson.M{"accessionnumber": md.AccessionNumber}
			update := bson.M{"$setOnInsert": md}
			opts := options.Update().SetUpsert(true)

			_, err := collection.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				errorChannel <- fmt.Errorf("failed to store metadata: %v", err)
				return
			}
		}(metaData)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Close the error channel as all goroutines have finished
	close(errorChannel)

	// Check for errors from goroutines
	for err := range errorChannel {
		if err != nil {
			return err // Return the first error encountered
		}
	}

	fmt.Println("stored metadata to Mongo filingMetaData")

	return nil
}

func Get10K10QMetadataFromSubmissionFilesGivenCIK(CIK string) ([]FilingMetaData, error) {
	submissionFilePaths, err := FindSubmissionFilesGivenCIK(CIK)
	if err != nil {
		return nil, err
	}
	var metadataSlice []FilingMetaData
	for _, submissionFilePath := range submissionFilePaths {
		if data, err := Extract10K10QmetadataFromSubmissionFile(submissionFilePath, CIK); err == nil {
			metadataSlice = append(metadataSlice, data...)
		} else {
			// Handle the error, e.g., log it or return it
			return nil, err
		}
	}
	return metadataSlice, nil
}

// Extract10K10QmetadataFromSubmissionFile processes a submission JSON file and extracts metadata
// for 10-K and 10-Q filings. It handles two types of JSON structures:
// 1. Recent filings JSON (uses "filings.recent." prefix)
// 2. Submissions JSON (uses no prefix)
//
// Parameters:
//   - filePath: Path to the JSON file containing SEC filing metadata
//   - CIK: Company Identifier Key for the company
//
// Returns:
//   - []FilingMetaData: Slice of metadata for all 10-K and 10-Q filings found
//   - error: Any error encountered during processing
//
// The function performs the following steps:
// 1. Reads and validates the JSON file
// 2. Determines the JSON structure type (recent filings vs submissions)
// 3. Finds all 10-K and 10-Q filings in the JSON
// 4. Extracts metadata for each filing (dates, numbers, etc.)
func Extract10K10QmetadataFromSubmissionFile(filePath string, CIK string) ([]FilingMetaData, error) {
	// Read and validate JSON file
	jsonString, err := ReadJsonFile(filePath)
	if err != nil {
		err = fmt.Errorf("couldn't read json file %v: %v", filePath, err)
		return nil, err
	}
	if !gjson.Valid(jsonString) {
		err = fmt.Errorf("invalid json %v", filePath)
		return nil, err
	}

	// Determine JSON structure type and set appropriate prefix
	var fileName string = filepath.Base(filePath)
	var doesFileNameIncludeSubmissions bool = strings.Contains(fileName, "submissions")
	var prefix string = "filings.recent."
	if doesFileNameIncludeSubmissions {
		prefix = ""
	}

	// Find indices of all 10-K and 10-Q filings in the JSON
	var IndiciesOf10K10Q []int
	listOfForms := gjson.Get(jsonString, prefix+"form")
	listOfForms.ForEach(func(key, value gjson.Result) bool {
		index := int(key.Int())
		form := value.String()

		if form == "10-K" || form == "10-Q" {
			IndiciesOf10K10Q = append(IndiciesOf10K10Q, index)
		}
		return true // Continue iterating over all items
	})

	// Extract metadata for each 10-K and 10-Q filing
	var FilingMetaDatSlice []FilingMetaData
	for i := 0; i < len(IndiciesOf10K10Q); i++ {
		// Create new metadata object for each filing
		var metaData FilingMetaData
		metaData.CIK = CIK

		// Extract all metadata fields using the filing's index
		// Each field is accessed using the appropriate JSON path with prefix
		metaData.AccessionNumber = gjson.Get(jsonString, fmt.Sprintf("%saccessionNumber.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.FilingDate = gjson.Get(jsonString, fmt.Sprintf("%sfilingDate.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.ReportDate = gjson.Get(jsonString, fmt.Sprintf("%sreportDate.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.AcceptanceDateTime = gjson.Get(jsonString, fmt.Sprintf("%sacceptanceDateTime.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.Act = gjson.Get(jsonString, fmt.Sprintf("%sact.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.Form = gjson.Get(jsonString, fmt.Sprintf("%sform.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.FileNumber = gjson.Get(jsonString, fmt.Sprintf("%sfileNumber.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.FilmNumber = gjson.Get(jsonString, fmt.Sprintf("%sfilmNumber.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.Items = gjson.Get(jsonString, fmt.Sprintf("%sitems.%d", prefix, IndiciesOf10K10Q[i])).String()
		metaData.Size = gjson.Get(jsonString, fmt.Sprintf("%ssize.%d", prefix, IndiciesOf10K10Q[i])).String()

		// Add the metadata to our collection
		FilingMetaDatSlice = append(FilingMetaDatSlice, metaData)
	}
	fmt.Println(FilingMetaDatSlice)

	return FilingMetaDatSlice, nil
}

// GetSubmissionFilesOfCIK returns a list of submission files for a given CIK
// the submission files need to be downloaded from SEC (https://www.sec.gov/search-filings/edgar-application-programming-interfaces)
// This function generates a list of all the submission files for a given CIK
func FindSubmissionFilesGivenCIK(CIK string) (submissionFilePaths []string, err error) {
	var submissionFiles []string
	// baseDirectory := "SEC-files/submissions"
	baseDirectory := filepath.Join("SEC-files", "submissions")

	files, err := os.ReadDir(baseDirectory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), CIK) {
			fullPath := filepath.Join(baseDirectory, file.Name())
			submissionFiles = append(submissionFiles, fullPath)
		}
	}

	return submissionFiles, nil
}

func ReadJsonFile(filePath string) (string, error) {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading JSON file: %v", err)
	}

	jsonString := string(jsonData)

	return jsonString, nil
}
