package fetchdata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	Size               int64  `json:"size"`
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

// func ParseOneSubmissionFile(filePath string) {

// }

func ReadJsonFile(filePath string) (string, error) {
	// Read the file content
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading JSON file: %v", err)
	}

	jsonString := string(jsonData)

	return jsonString, nil
}
