package combinecsvfiles

import (
	"errors"
	"fmt"

	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
)

// This function finds the common fields for all financial statement rfiles and returns them
// FirstEmptyRow serves as the separation row between header cells and data cells of R files that have been parsed
func CommonFieldFinderForFinancialStatementRfile(RfilePath string) (reportDate_ string, form_ string, accessionNumber_ string, totalLineItemCount_ int, separatorRow_ int, err error) {
	RfileData2Darray, err := utilityfunctions.ReadCsvFileToArray(RfilePath)
	if err != nil {
		return "", "", "", 0, 0, err
	}

	if len(RfileData2Darray) == 0 {
		return "", "", "", 0, 0, errors.New("empty CSV file")
	}

	var reportDate string
	var form string
	var accessionNumber string
	var totalLineItemCount int
	var separatorRow int

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
		return "", "", "", 0, 0, errors.New("reportDate not found")
	}
	if form == "" {
		return "", "", "", 0, 0, errors.New("form not found")
	}
	if accessionNumber == "" {
		return "", "", "", 0, 0, errors.New("accessionNumber not found")
	}

	//find first separator row
	for i := 0; i < len(RfileData2Darray); i++ {
		if RfileData2Darray[i][0] == "separator" {
			separatorRow = i
			break
		}
	}

	// below separator row check for totalLineItemCount
	// first check if first column is not empty, then check if any other cell in that row is not empty
	columnCount := len(RfileData2Darray[0])
	for i := separatorRow + 1; i < len(RfileData2Darray); i++ {
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
	fmt.Println("separatorRow: ", separatorRow)

	return reportDate, form, accessionNumber, totalLineItemCount, separatorRow, nil
}
