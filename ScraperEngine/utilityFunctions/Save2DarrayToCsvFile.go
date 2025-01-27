// takes in 2D array and output directory and filename and saves it as a csv file
package utilityfunctions

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

func Save2DarrayToCsvFile(array [][]string, directory string, filename string) error {
	//create the directory if it does not exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}
	//create the file
	file, err := os.Create(filepath.Join(directory, filename))
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()
	//write the array to the file
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(array); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
