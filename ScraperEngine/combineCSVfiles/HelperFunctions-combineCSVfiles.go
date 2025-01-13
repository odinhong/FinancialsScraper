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

// CheckWordsInOrder checks if three words exist in a string in the given order without overlapping
func CheckWordsInOrder(text string, words []string) bool {
	if len(words) != 3 {
		return false
	}

	currentIndex := 0
	for _, word := range words {
		// Find the word starting from the current index
		index := strings.Index(text[currentIndex:], word)
		if index == -1 {
			return false
		}
		// Move the current index past the found word
		currentIndex += index + len(word)
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

// CompareIntSlices checks if two int slices contain the same elements regardless of order
func CompareIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	
	// Create maps to count occurrences
	countA := make(map[int]int)
	countB := make(map[int]int)
	
	// Count occurrences in both slices
	for _, val := range a {
		countA[val]++
	}
	for _, val := range b {
		countB[val]++
	}
	
	// Compare the counts
	for key, count := range countA {
		if countB[key] != count {
			return false
		}
	}
	
	return true
}
