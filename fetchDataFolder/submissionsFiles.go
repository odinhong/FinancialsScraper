package fetchdata

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	metadataSlice, err := Get10K10QMetadataFromSubmissionFilesCIK(CIK)
	if err != nil {
		return err
	}
	// fmt.Println("metadataSlice:", metadataSlice)

	databaseName := os.Getenv("DATABASE_NAME")
	collectionName := os.Getenv("COLLECTION_NAME")
	// databaseName := "testDatabase"
	// collectionName := "testMetaDataOf10K10Q"
	collection := client.Database(databaseName).Collection(collectionName)

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
func Get10K10QMetadataFromSubmissionFilesCIK(CIK string) ([]FilingMetaData, error) {
	submissionFiles, err := GetSubmissionFilesOfCIK(CIK)
	if err != nil {
		return nil, err
	}
	var metadataSlice []FilingMetaData
	for _, submissionFile := range submissionFiles {
		if data, err := Parse10K10QmetadataFromSubmissionJsonFile(submissionFile, CIK); err == nil {
			metadataSlice = append(metadataSlice, data...)
		} else {
			// Handle the error, e.g., log it or return it
			return nil, err
		}
	}
	return metadataSlice, nil
}

func Parse10K10QmetadataFromSubmissionJsonFile(filePath string, CIK string) ([]FilingMetaData, error) {
	jsonString, err := ReadJsonFile(filePath)
	if err != nil {
		err = fmt.Errorf("couldn't read json file %v: %v", filePath, err)
		return nil, err
	}
	if !gjson.Valid(jsonString) {
		err = fmt.Errorf("invalid json %v", filePath)
		return nil, err
	}

	var fileName string = filepath.Base(filePath)
	var doesFileNameIncludeSubmissions bool = strings.Contains(fileName, "submissions")

	var prefix string = "filings.recent."
	if doesFileNameIncludeSubmissions {
		prefix = ""
	}

	var locationsOf10K10Q []int
	listOfForms := gjson.Get(jsonString, prefix+"form")
	listOfForms.ForEach(func(key, value gjson.Result) bool {
		index := int(key.Int())
		form := value.String()

		if form == "10-K" || form == "10-Q" {
			locationsOf10K10Q = append(locationsOf10K10Q, index)
		}
		return true // Continue iterating over all items
	})

	var FilingMetaDatSlice []FilingMetaData
	for i := 0; i < len(locationsOf10K10Q); i++ {
		var metaData FilingMetaData
		metaData.CIK = CIK
		metaData.AccessionNumber = gjson.Get(jsonString, fmt.Sprintf("%saccessionNumber.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.FilingDate = gjson.Get(jsonString, fmt.Sprintf("%sfilingDate.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.ReportDate = gjson.Get(jsonString, fmt.Sprintf("%sreportDate.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.AcceptanceDateTime = gjson.Get(jsonString, fmt.Sprintf("%sacceptanceDateTime.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.Act = gjson.Get(jsonString, fmt.Sprintf("%sact.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.Form = gjson.Get(jsonString, fmt.Sprintf("%sform.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.FileNumber = gjson.Get(jsonString, fmt.Sprintf("%sfileNumber.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.FilmNumber = gjson.Get(jsonString, fmt.Sprintf("%sfilmNumber.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.Items = gjson.Get(jsonString, fmt.Sprintf("%sitems.%d", prefix, locationsOf10K10Q[i])).String()
		metaData.Size = gjson.Get(jsonString, fmt.Sprintf("%ssize.%d", prefix, locationsOf10K10Q[i])).String()

		FilingMetaDatSlice = append(FilingMetaDatSlice, metaData)
	}
	fmt.Println(FilingMetaDatSlice)

	return FilingMetaDatSlice, nil
}

// GetSubmissionFilesOfCIK returns a list of submission files for a given CIK
// the submission files need to be downloaded from SEC (https://www.sec.gov/search-filings/edgar-application-programming-interfaces)
// This function generates a list of all the submission files for a given CIK
func FindSubmissionFiles(CIK string) (submissionFilePaths []string, err error) {
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
