package combinecsvfiles

import (
	"fmt"
	"strings"
)

// Helper function to check if any string in the array is contained in the target
func containsAny(target string, possibilities []string) bool {
	// Remove spaces and convert to lowercase
	target = strings.ToLower(strings.ReplaceAll(target, " ", ""))
	for _, possible := range possibilities {
		// Remove spaces and convert to lowercase for comparison
		possible = strings.ToLower(strings.ReplaceAll(possible, " ", ""))
		if strings.Contains(target, possible) {
			return true
		}
	}
	return false
}

// Helper function to check if none of the strings in the array are contained in the target
func notContainsAny(target string, possibilities []string) bool {
	// Remove spaces and convert to lowercase
	target = strings.ToLower(strings.ReplaceAll(target, " ", ""))
	for _, possible := range possibilities {
		// Remove spaces and convert to lowercase for comparison
		possible = strings.ToLower(strings.ReplaceAll(possible, " ", ""))
		if strings.Contains(target, possible) {
			return false
		}
	}
	return true
}

// CheckWordsInOrder checks if phrases exist in a string in the given order without overlapping
func CheckWordsInOrder(text string, phrases []string) bool {
	if len(phrases) == 0 {
		return false
	}

	// Convert text to lowercase and trim spaces
	text = strings.ToLower(strings.TrimSpace(text))
	currentIndex := 0

	for _, phrase := range phrases {
		// Convert phrase to lowercase and trim spaces
		phrase = strings.ToLower(strings.TrimSpace(phrase))

		// Find the phrase starting from the current index
		index := strings.Index(text[currentIndex:], phrase)
		if index == -1 {
			return false
		}
		// Move the current index past the found phrase
		currentIndex += index + len(phrase)
	}
	return true
}

// Helper function to check if the row contains a data cell that is not empty
func DoesDataCellExistInThisRow(row []string) bool {
	if len(row) == 0 || row[0] == "" {
		return false
	}
	// Check if any column after the first has data
	for i := 1; i < len(row); i++ {
		if row[i] != "" {
			return true
		}
	}
	return false
}

// CheckBalanceSheetOrder checks if the line items are in the correct order
func CheckBalanceSheetOrder(indices map[string]int) error {
	// Define the expected order of sections
	expectedOrder := []string{
		"currentAssets",
		"totalCurrentAssets",
		"totalAssets",
		"currentLiabilities",
		"totalCurrentLiabilities",
		"totalLiabilities",
		"stockholdersEquity",
		"totalStockholdersEquity",
	}

	// Check each pair of adjacent sections
	for i := 0; i < len(expectedOrder)-1; i++ {
		current := indices[expectedOrder[i]]
		next := indices[expectedOrder[i+1]]

		if current == -1 || next == -1 {
			return fmt.Errorf("missing required section: %s or %s", expectedOrder[i], expectedOrder[i+1])
		}

		if current > next {
			return fmt.Errorf("incorrect order: %s (row %d) should come before %s (row %d)",
				expectedOrder[i], current, expectedOrder[i+1], next)
		}
	}

	return nil
}
