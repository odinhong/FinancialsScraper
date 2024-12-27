package combinecsvfiles

import (
	"errors"
	"fmt"
	"strings"

	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
)

// process as much as possible with code. For line items code cant deal with, ChatGPT api will be needed
func ProcessBalanceSheetCsvRfile(RfilePath string) {
	RfileData2Darray, reportDate, form, accessionNumber, totalLineItemCount, separatorRowIndex, err := CommonInitialProcessorForCsvRfile(RfilePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	//find the row index for Assets, Liabilities, Equity, other equity, and Other
	//Sub divide Assets, into CurrentAssets, NonCurrentAssets
	//Sub divide Liabilities, into CurrentLiabilities, NonCurrentLiabilities

	var startRowIndex_Assets int
	var endRowIndex_Assets int
	var startRowIndex_Liabilities int
	var endRowIndex_Liabilities int
	var startRowIndex_Equity int
	var endRowIndex_Equity int
	var startRowIndex_OtherEquity int
	var endRowIndex_OtherEquity int

	var startRowIndex_CurrentAssets int
	var endRowIndex_CurrentAssets int
	var startRowIndex_NonCurrentAssets int
	var endRowIndex_NonCurrentAssets int

	var startRowIndex_CurrentLiabilities int
	var endRowIndex_CurrentLiabilities int
	var startRowIndex_NonCurrentLiabilities int
	var endRowIndex_NonCurrentLiabilities int

	possibleWords_TotalAssets := []string{"Total Assets"}
	possibleWords_TotalLiabilities := []string{"Total Liabilities"}
	possibleWords_TotalEquity := []string{"Total Stockholders"}
	possibleWords_OtherEquity := []string{"Preferred Stock", "Preferred Equity", "Convertible Debt"}

	possibleWords_CurrentAssets := []string{"Total Current Assets"}
	possibleWords_CurrentLiabilities := []string{"Total Current Liabilities"}

	startRowIndex_Assets = separatorRowIndex + 1 //Assuming balance sheet starts after separator row and starts with Assets

	type BalanceSheetSection struct {
		rowIndex    *int
		searchWords []string
		found       bool
	}

	BalanceSheetSections := []BalanceSheetSection{
		{&endRowIndex_Assets, possibleWords_TotalAssets, false},
		{&endRowIndex_Liabilities, possibleWords_TotalLiabilities, false},
		{&endRowIndex_Equity, possibleWords_TotalEquity, false},
		{&endRowIndex_CurrentAssets, possibleWords_CurrentAssets, false},
		{&endRowIndex_CurrentLiabilities, possibleWords_CurrentLiabilities, false},
	}

	for i := separatorRowIndex + 1; i < len(RfileData2Darray) && DoesDataCellExistInThisRow(RfileData2Darray[i]); i++ {
		for j := range BalanceSheetSections {
			if !BalanceSheetSections[j].found && containsAny(RfileData2Darray[i][0], BalanceSheetSections[j].searchWords) {
				*BalanceSheetSections[j].rowIndex = i
				BalanceSheetSections[j].found = true
			}
		}
	}

	startRowIndex_Assets = separatorRowIndex + 1
	startRowIndex_Liabilities = endRowIndex_Assets + 1
	startRowIndex_Equity = endRowIndex_Liabilities + 1
	startRowIndex_CurrentAssets = separatorRowIndex + 1
	startRowIndex_CurrentLiabilities = endRowIndex_Assets + 1

	endRowIndex_NonCurrentAssets = endRowIndex_Assets
	endRowIndex_NonCurrentLiabilities = endRowIndex_Liabilities

}

// This function finds the common fields for all financial statement rfiles and returns them
// FirstEmptyRow serves as the separation row between header cells and data cells of R files that have been parsed
func CommonInitialProcessorForCsvRfile(RfilePath string) (RfileData2Darray_ [][]string, reportDate_ string, form_ string, accessionNumber_ string, totalLineItemCount_ int, separatorRowIndex_ int, err error) {
	RfileData2Darray, err := utilityfunctions.ReadCsvFileToArray(RfilePath)
	if err != nil {
		return nil, "", "", "", 0, 0, err
	}

	if len(RfileData2Darray) == 0 {
		return nil, "", "", "", 0, 0, errors.New("empty CSV file")
	}

	var reportDate string
	var form string
	var accessionNumber string
	var totalLineItemCount int
	var separatorRowIndex int

	//check first column of first 10 rows for reportDate, form, accessionNumber
	for i := 0; i < 10; i++ {
		if RfileData2Darray[i][0] == "reportDate" {
			reportDate = RfileData2Darray[i][1]
		}
		if RfileData2Darray[i][0] == "form" {
			form = RfileData2Darray[i][1]
		}
		if RfileData2Darray[i][0] == "accessionNumber" {
			accessionNumber = RfileData2Darray[i][1]
		}
	}

	//return error if reportDate, form, or accessionNumber not found
	if reportDate == "" {
		return nil, "", "", "", 0, 0, errors.New("reportDate not found")
	}
	if form == "" {
		return nil, "", "", "", 0, 0, errors.New("form not found")
	}
	if accessionNumber == "" {
		return nil, "", "", "", 0, 0, errors.New("accessionNumber not found")
	}

	//find first separator row
	for i := 0; i < len(RfileData2Darray); i++ {
		if RfileData2Darray[i][0] == "separator" {
			separatorRowIndex = i
			break
		}
	}

	// below separator row check for totalLineItemCount
	// first check if first column is not empty, then check if any other cell in that row is not empty
	columnCount := len(RfileData2Darray[0])
	for i := separatorRowIndex + 1; i < len(RfileData2Darray); i++ {
		if RfileData2Darray[i][0] != "" { // check first column
			// if first column not empty, check other columns
			for j := 1; j < columnCount; j++ {
				if RfileData2Darray[i][j] != "" {
					totalLineItemCount++
					break // found a non-empty cell, no need to check rest of row
				}
			}
		}
	}

	fmt.Println("reportDate: ", reportDate)
	fmt.Println("form: ", form)
	fmt.Println("accessionNumber: ", accessionNumber)
	fmt.Println("totalLineItemCount: ", totalLineItemCount)
	fmt.Println("separatorRowIndex: ", separatorRowIndex)

	return RfileData2Darray, reportDate, form, accessionNumber, totalLineItemCount, separatorRowIndex, nil
}

// Helper function to check if any string in the array is contained in the target
func containsAny(target string, possibilities []string) bool {
	for _, possible := range possibilities {
		if strings.Contains(target, possible) {
			return true
		}
	}
	return false
}

// Helper function to check if the row contains a data cell that is not empty
func DoesDataCellExistInThisRow(row []string) bool {
	for i := 1; i < len(row); i++ {
		if row[0] != "" && row[i] != "" {
			return true
		}
	}
	return false
}
