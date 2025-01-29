package combinecsvfiles

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func FillInDataCellsToEmptyCombinedStatement(emptyCombinedStatement [][]string, OriginalStatement1 [][]string, OriginalStatement2 [][]string) (FilledCombinedStatement [][]string, err error) {
	//use the header cells and metadata cells to essentially do a lookup on the original statements to fill in the data cells

	//find all required row indices
	separatorRowIndexStatement1 := -1
	accessionNumberRowIndexStatement1 := -1
	reportDateRowIndexStatement1 := -1
	reportPeriodRowIndexStatement1 := -1
	separatorRowIndexStatement2 := -1
	accessionNumberRowIndexStatement2 := -1
	reportDateRowIndexStatement2 := -1
	reportPeriodRowIndexStatement2 := -1

	for i := 0; i < len(OriginalStatement1); i++ {
		// Clean the string by removing BOM and trimming whitespace
		cleanedString := removeBOM(strings.TrimSpace(OriginalStatement1[i][0]))
		switch cleanedString {
		case "separator":
			separatorRowIndexStatement1 = i
		case "accessionNumber":
			accessionNumberRowIndexStatement1 = i
		case "reportDate":
			reportDateRowIndexStatement1 = i
		case "reportPeriod":
			reportPeriodRowIndexStatement1 = i
		}
	}

	for i := 0; i < len(OriginalStatement2); i++ {
		// Clean the string by removing BOM and trimming whitespace
		cleanedString := removeBOM(strings.TrimSpace(OriginalStatement2[i][0]))
		switch cleanedString {
		case "separator":
			separatorRowIndexStatement2 = i
		case "accessionNumber":
			accessionNumberRowIndexStatement2 = i
		case "reportDate":
			reportDateRowIndexStatement2 = i
		case "reportPeriod":
			reportPeriodRowIndexStatement2 = i
		}
	}

	//check if all the indices are found
	if separatorRowIndexStatement1 == -1 || separatorRowIndexStatement2 == -1 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: separator row not found")
		return emptyCombinedStatement, err
	}
	if accessionNumberRowIndexStatement1 == -1 || accessionNumberRowIndexStatement2 == -1 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: accessionNumber row not found")
		return emptyCombinedStatement, err
	}
	if reportDateRowIndexStatement1 == -1 || reportDateRowIndexStatement2 == -1 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: reportDate row not found")
		return emptyCombinedStatement, err
	}
	if reportPeriodRowIndexStatement1 == -1 || reportPeriodRowIndexStatement2 == -1 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: reportPeriod row not found")
		return emptyCombinedStatement, err
	}

	//check if the report dates and report periods match
	if separatorRowIndexStatement1 != separatorRowIndexStatement2 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: separator row indices do not match")
		return emptyCombinedStatement, err
	}
	if accessionNumberRowIndexStatement1 != accessionNumberRowIndexStatement2 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: accessionNumber row indices do not match")
		return emptyCombinedStatement, err
	}
	if reportDateRowIndexStatement1 != reportDateRowIndexStatement2 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: reportDate row indices do not match")
		return emptyCombinedStatement, err
	}
	if reportPeriodRowIndexStatement1 != reportPeriodRowIndexStatement2 {
		err := errors.New("function FillInDataCellsToEmptyCombinedStatement: reportPeriod row indices do not match")
		return emptyCombinedStatement, err
	}

	//fill in the data cells one by one
	FilledCombinedStatement = emptyCombinedStatement
	for i := separatorRowIndexStatement1 + 1; i < len(emptyCombinedStatement); i++ {
		for j := 1; j < len(emptyCombinedStatement[0]); j++ {
			lineItemName := emptyCombinedStatement[i][0]
			accessionNumber := emptyCombinedStatement[accessionNumberRowIndexStatement1][j]
			reportDate := emptyCombinedStatement[reportDateRowIndexStatement1][j]
			reportPeriod := emptyCombinedStatement[reportPeriodRowIndexStatement1][j]
			//check first statement to find the cellvalue, if not found, check the second statement
			cellValue, err := LookupCellValueGivenHeaderAndMetadataCells(OriginalStatement1, lineItemName, accessionNumber, accessionNumberRowIndexStatement1, reportPeriod, reportPeriodRowIndexStatement1, reportDate, reportDateRowIndexStatement1)
			if cellValue == "" && err != nil {
				cellValue, _ = LookupCellValueGivenHeaderAndMetadataCells(OriginalStatement2, lineItemName, accessionNumber, accessionNumberRowIndexStatement2, reportPeriod, reportPeriodRowIndexStatement2, reportDate, reportDateRowIndexStatement2)
			}

			FilledCombinedStatement[i][j] = cellValue
		}
	}

	return FilledCombinedStatement, nil
}

func LookupCellValueGivenHeaderAndMetadataCells(OriginalStatement [][]string, lineItemName string, accessionNumber string, accessionNumberRowIndex int, reportPeriod string, reportPeriodRowIndex int, reportDate string, reportDateRowIndex int) (cellValue string, err error) {
	rowIndex := -1
	columnIndex := -1

	// Validate input parameters
	if len(OriginalStatement) == 0 || len(OriginalStatement[0]) == 0 {
		return "", errors.New("empty statement array")
	}

	// Helper function to normalize strings for comparison
	normalizeString := func(s string) string {
		// Convert to lowercase and remove all spaces
		return strings.ReplaceAll(strings.ToLower(s), " ", "")
	}

	for i := 1; i < len(OriginalStatement); i++ {
		if normalizeString(OriginalStatement[i][0]) == normalizeString(lineItemName) {
			rowIndex = i
			break
		}
	}

	//go to each col and see which col meets all the conditions
	for i := 1; i < len(OriginalStatement[0]); i++ {
		if OriginalStatement[accessionNumberRowIndex][i] == accessionNumber {
			if OriginalStatement[reportDateRowIndex][i] == reportDate {
				if OriginalStatement[reportPeriodRowIndex][i] == reportPeriod {
					columnIndex = i
					break
				}
			}
		}
	}

	if rowIndex == -1 || columnIndex == -1 {
		return "", errors.New("could not locate cell")
	}

	// Additional bounds check before accessing the array
	if rowIndex >= len(OriginalStatement) || columnIndex >= len(OriginalStatement[rowIndex]) {
		return "", errors.New("index out of range")
	}

	//clean the cell value that look like this: -228us-gaap_AccumulatedOtherComprehensiveIncomeLossNetOfTax from https://www.sec.gov/Archives/edgar/data/1326801/000132680115000006 click on one of the line item value in the link to see the text
	cellValue = strings.TrimSpace(OriginalStatement[rowIndex][columnIndex])
	// Regular expression to match a number (including negative) at the start of the string (after any whitespace)
	re := regexp.MustCompile(`^\s*(-?\d+)`)
	matches := re.FindStringSubmatch(cellValue)
	if len(matches) > 1 {
		// matches[1] contains the captured number
		cellValue = matches[1]
	}

	return cellValue, nil
}

// removeBOM removes the UTF-8 Byte Order Mark from the beginning of a string if present
func removeBOM(s string) string {
	if strings.HasPrefix(s, "\ufeff") {
		return strings.TrimPrefix(s, "\ufeff")
	}
	return s
}

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

// RearrangeAllColumns rearranges columns based on the desired indices, preserving the first column for headers
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

// DeleteColumn removes a column at the specified index from a 2D array
func DeleteColumn(array [][]string, colIndex int) [][]string {
	result := make([][]string, len(array))
	for i := range array {
		// Create a new row without the specified column
		result[i] = append(array[i][:colIndex], array[i][colIndex+1:]...)
	}
	return result
}

// DeleteRow removes a row at the specified index from a 2D array
func DeleteRow(array [][]string, rowIndex int) [][]string {
	result := make([][]string, len(array))
	for i := range array {
		if i == rowIndex {
			continue
		}
		result[i] = array[i]
	}
	return result
}

// Helper function to check if any string in the array is contained in the target
// containsAny is case and space insensitive
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

// DoesDataCellExistInThisRow Helper function to check if the row contains a data cell that is not empty
func DoesDataCellExistInThisRow(row []string) bool {
	if len(row) == 0 || row[0] == "" {
		return false
	}
	// Check if any column after the first cell has data
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
