package fetchdata

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FilingMetaData struct {
	AccessionNumber    string `json:"accessionNumber"`
	FilingDate         string `json:"filingDate"`
	ReportDate         string `json:"reportDate"`
	AcceptanceDateTime string `json:"acceptanceDateTime"`
	Act                string `json:"act"`
	Form               string `json:"form"`
	FileNumber         string `json:"fileNumber"`
	FilmNumber         string `json:"filmNumber"`
	Items              string `json:"items"`
	Size               string `json:"size"`
}

// var filingMetaDataCollection *mongo.Collection

func Store10K10QmetadataFromSubmissionFilesCIKtoMongoDB(client *mongo.Client, CIK string) error {
	metadataSlice, err := Get10K10QMetadataFromSubmissionFilesCIK(CIK)
	if err != nil {
		return err
	}

	databaseName := "testDatabase"
	collectionName := "testMetaDataOf10K10Q"
	collection := client.Database(databaseName).Collection(collectionName)

	// Use a context with timeout or a background context as needed
	ctx := context.Background()

	// Insert or update each meta data record into MongoDB
	for _, metaData := range metadataSlice {
		filter := bson.M{"accessionNumber": metaData.AccessionNumber}
		update := bson.M{"$setOnInsert": metaData}
		opts := options.Update().SetUpsert(true)

		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("failed to store metadata: %v", err)
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
		if data, err := Parse10K10QmetadataFromSubmissionJsonFile(submissionFile); err == nil {
			metadataSlice = append(metadataSlice, data...)
		} else {
			// Handle the error, e.g., log it or return it
			return nil, err
		}
	}
	return metadataSlice, nil
}

func Parse10K10QmetadataFromSubmissionJsonFile(filePath string) ([]FilingMetaData, error) {
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

	return FilingMetaDatSlice, nil
}

func GetSubmissionFilesOfCIK(CIK string) ([]string, error) {
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
