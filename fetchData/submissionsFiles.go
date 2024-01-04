package fetchdata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
)

type Filing struct {
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

func ParseOneJsonSubmissionFile(filePath string) ([]int, error) {
	jsonString, err := ReadJsonFile(filePath)
	if err != nil {
		err = fmt.Errorf("couldn't read json file %v: %v", filePath, err)
		return nil, err
	}
	if !gjson.Valid(jsonString) {
		err = fmt.Errorf("invalid json %v", filePath)
		return nil, err
	}

	var fileName string = strings.ReplaceAll(filePath, "SEC-files/submissions/", "")
	var doesFileNameIncludeSubmissions bool = strings.Contains(fileName, "submissions")

	var prefix string = "filings.recent."
	if doesFileNameIncludeSubmissions {
		prefix = ""
	}

	var locationsOf10K10Q []int
	testing := gjson.Get(jsonString, prefix+"form")
	testing.ForEach(func(key, value gjson.Result) bool {
		index := int(key.Int())
		form := value.String()

		if form == "10-K" || form == "10-Q" {
			locationsOf10K10Q = append(locationsOf10K10Q, index)
		}
		return true // Continue iterating over all items
	})

	return locationsOf10K10Q, nil
}

func GetSubmissionFilesOfCIK(CIK string) ([]string, error) {
	var submissionFiles []string
	baseDirectory := "SEC-files/submissions"

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
	// Read the file content
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading JSON file: %v", err)
	}

	jsonString := string(jsonData)

	return jsonString, nil
}
