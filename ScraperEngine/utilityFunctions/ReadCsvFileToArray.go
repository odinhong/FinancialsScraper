package utilityfunctions

import (
	"encoding/csv"
	"os"
)

func ReadCsvFileToArray(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Tell the reader to accept variable number of fields per row
	reader.FieldsPerRecord = -1

	// Read all records at once
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}
