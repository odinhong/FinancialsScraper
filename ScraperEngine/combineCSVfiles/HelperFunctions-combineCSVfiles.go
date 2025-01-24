package combinecsvfiles

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ColInfo holds information about a column for sorting
type ColInfo struct {
	index        int
	reportPeriod int
	reportDate   int
}

// GetIndexOfRearrangedColumnsByReportPeriodAndDate returns indices of columns sorted by reportPeriod (ascending) and then reportDate (ascending)
func GetIndexOfRearrangedColumnsByReportPeriodAndDate(data [][]string, reportPeriodRowIndex, reportDateRowIndex int) []int {
	if len(data) == 0 {
		return nil
	}

	// Create slice of column info
	colInfos := make([]ColInfo, len(data[0])-1) // -1 because first column is headers
	for i := 1; i < len(data[0]); i++ {
		// Convert strings to integers, default to 0 if conversion fails
		reportPeriod, _ := strconv.Atoi(data[reportPeriodRowIndex][i])
		reportDate, _ := strconv.Atoi(data[reportDateRowIndex][i])

		colInfos[i-1] = ColInfo{
			index:        i,
			reportPeriod: reportPeriod,
			reportDate:   reportDate,
		}
	}

	// Sort the columns based on reportPeriod and reportDate
	sort.Slice(colInfos, func(i, j int) bool {
		// If reportPeriods are different, sort by reportPeriod
		if colInfos[i].reportPeriod != colInfos[j].reportPeriod {
			return colInfos[i].reportPeriod < colInfos[j].reportPeriod
		}
		// If reportPeriods are same, sort by reportDate
		return colInfos[i].reportDate < colInfos[j].reportDate
	})

	// Extract the sorted indices
	result := make([]int, len(colInfos))
	for i, info := range colInfos {
		result[i] = info.index
	}

	return result
}

// RearrangeAllColumns rearranges columns based on the desired indices
// columnOrder contains indices for data columns (1 onwards)
func RearrangeAllColumns(data [][]string, columnOrder []int) [][]string {
	if len(data) == 0 {
		return data
	}

	result := make([][]string, len(data))

	for i := range data {
		row := make([]string, len(data[i]))
		// Keep first column
		row[0] = data[i][0]

		// Rearrange data columns (shifted by 1 since first column is preserved)
		for newPos, oldPos := range columnOrder {
			row[newPos+1] = data[i][oldPos]
		}

		result[i] = row
	}

	return result
}

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

// CheckIfLineItemNameIsInLineItemNameList checks if a line item name is in the list
func CheckIfLineItemNameIsInLineItemNameList(lineItemName string, lineItemNameList []string) bool {
	for _, item := range lineItemNameList {
		if lineItemName == item {
			return true
		}
	}
	return false
}
