package combinecsvfiles

import (
	"errors"
	"fmt"
	"strings"

	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
)

// process as much as possible with code. For line items code cant deal with, ChatGPT api will be needed
type BalanceSheetIndices struct {
	ReportDate                 string
	Form                       string
	AccessionNumber            string
	TotalLineItemCount         int
	SeparatorRowIndex          int
	AssetsStart                int
	AssetsEnd                  int
	LiabilitiesStart           int
	LiabilitiesEnd             int
	EquityStart                int
	EquityEnd                  int
	OtherEquityStart           int
	OtherEquityEnd             int
	CurrentAssetsStart         int
	CurrentAssetsEnd           int
	NonCurrentAssetsStart      int
	NonCurrentAssetsEnd        int
	CurrentLiabilitiesStart    int
	CurrentLiabilitiesEnd      int
	NonCurrentLiabilitiesStart int
	NonCurrentLiabilitiesEnd   int
}

func ProcessBalanceSheetCsvRfile(RfilePath string) (BalanceSheetIndices, error) {
	RfileData2Darray, reportDate, form, accessionNumber, totalLineItemCount, separatorRowIndex, err := CommonInitialProcessorForCsvRfile(RfilePath)
	if err != nil {
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	//find the row index for Assets, Liabilities, Equity, other equity, and Other
	//Sub divide Assets, into CurrentAssets, NonCurrentAssets
	//Sub divide Liabilities, into CurrentLiabilities, NonCurrentLiabilities

	//int variables are set to 0 by default
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
	possibleWords_commonStock := []string{"Common Stock", "Common Share"}
	possibleWords_TotalEquity := []string{"Total Stockholders", "Total Shareholders"}
	possibleWords_OtherEquity := []string{"Preferred Stock", "Preferred Equity", "Convertible Debt"}

	possibleWords_CurrentAssets := []string{"Total Current Assets"}
	possibleWords_StartCurrentAssets := []string{"Cash and Cash Equivalents"}
	possibleWords_CurrentLiabilities := []string{"Total Current Liabilities"}

	type BalanceSheetSection struct {
		rowIndex    *int
		searchWords []string
		found       bool
	}

	BalanceSheetSections := []BalanceSheetSection{
		{&endRowIndex_Assets, possibleWords_TotalAssets, false},
		{&endRowIndex_Liabilities, possibleWords_TotalLiabilities, false},
		{&startRowIndex_Equity, possibleWords_commonStock, false},
		{&endRowIndex_Equity, possibleWords_TotalEquity, false},
		{&endRowIndex_CurrentAssets, possibleWords_CurrentAssets, false},
		{&startRowIndex_CurrentAssets, possibleWords_StartCurrentAssets, false},
		{&endRowIndex_CurrentLiabilities, possibleWords_CurrentLiabilities, false},
		{&startRowIndex_OtherEquity, possibleWords_OtherEquity, false},
	}

	for i := separatorRowIndex + 1; i < len(RfileData2Darray); i++ {
		if DoesDataCellExistInThisRow(RfileData2Darray[i]) {
			for j := range BalanceSheetSections {
				if !BalanceSheetSections[j].found {
					if containsAny(RfileData2Darray[i][0], BalanceSheetSections[j].searchWords) {
						*BalanceSheetSections[j].rowIndex = i
						BalanceSheetSections[j].found = true
					}
				}
			}
		}
	}

	startRowIndex_Assets = separatorRowIndex + 1 //Assuming balance sheet starts after separator row and starts with Assets
	startRowIndex_Liabilities = endRowIndex_Assets + 1
	startRowIndex_CurrentLiabilities = endRowIndex_Assets + 1
	startRowIndex_NonCurrentAssets = endRowIndex_CurrentAssets + 1
	startRowIndex_NonCurrentLiabilities = endRowIndex_CurrentLiabilities + 1

	endRowIndex_NonCurrentAssets = endRowIndex_Assets - 1
	endRowIndex_NonCurrentLiabilities = endRowIndex_Liabilities - 1

	//check if startRowIndex_OtherEquity is before startRowIndex_Equity
	//if true, set endRowIndex_OtherEquity to startRowIndex_Equity - 1
	if startRowIndex_OtherEquity < startRowIndex_Equity {
		endRowIndex_OtherEquity = startRowIndex_Equity - 1
	}
	//check if startRowIndex_OtherEquity is after startRowIndex_Equity && before endRowIndex_Equity
	//if true, return error and ChatGPT api will be needed to process this R file
	if startRowIndex_OtherEquity > startRowIndex_Equity && startRowIndex_OtherEquity < endRowIndex_Equity {
		err := errors.New("startRowIndex_OtherEquity is after startRowIndex_Equity and before endRowIndex_Equity")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	//check if there are data cells that are not accounted for by the sections described
	//if there are unaccounted data cells, ChatGPT api will be needed to process this R file
	for i := separatorRowIndex + 1; i < len(RfileData2Darray) && DoesDataCellExistInThisRow(RfileData2Darray[i]); i++ {
		// Skip if row is within Assets section
		if i >= startRowIndex_Assets && i <= endRowIndex_Assets {
			continue
		}
		// Skip if row is within Liabilities section
		if i >= startRowIndex_Liabilities && i <= endRowIndex_Liabilities {
			continue
		}
		// Skip if row is within Equity section
		if i >= startRowIndex_Equity && i <= endRowIndex_Equity {
			continue
		}
		// Skip if row is within Other Equity section (if it exists)
		if startRowIndex_OtherEquity != 0 && i >= startRowIndex_OtherEquity && i <= endRowIndex_OtherEquity {
			continue
		}
		// If we get here, we found a row with data that's not in any section
		err := errors.New("found unaccounted data cells outside of balance sheet sections")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	fmt.Println("startRowIndex_Assets", startRowIndex_Assets)
	fmt.Println("endRowIndex_Assets", endRowIndex_Assets)
	fmt.Println("startRowIndex_Liabilities", startRowIndex_Liabilities)
	fmt.Println("endRowIndex_Liabilities", endRowIndex_Liabilities)
	fmt.Println("startRowIndex_Equity", startRowIndex_Equity)
	fmt.Println("endRowIndex_Equity", endRowIndex_Equity)
	fmt.Println("startRowIndex_OtherEquity", startRowIndex_OtherEquity)
	fmt.Println("endRowIndex_OtherEquity", endRowIndex_OtherEquity)
	fmt.Println("startRowIndex_CurrentAssets", startRowIndex_CurrentAssets)
	fmt.Println("endRowIndex_CurrentAssets", endRowIndex_CurrentAssets)
	fmt.Println("startRowIndex_CurrentLiabilities", startRowIndex_CurrentLiabilities)
	fmt.Println("endRowIndex_CurrentLiabilities", endRowIndex_CurrentLiabilities)
	fmt.Println("startRowIndex_NonCurrentAssets", startRowIndex_NonCurrentAssets)
	fmt.Println("endRowIndex_NonCurrentAssets", endRowIndex_NonCurrentAssets)
	fmt.Println("startRowIndex_NonCurrentLiabilities", startRowIndex_NonCurrentLiabilities)
	fmt.Println("endRowIndex_NonCurrentLiabilities", endRowIndex_NonCurrentLiabilities)

	//check for inconsistency in row indexes of all balance sheet sections
	//check if any of the rows are above or equal to separatorRowIndex
	if startRowIndex_Assets <= separatorRowIndex ||
		startRowIndex_Liabilities <= separatorRowIndex ||
		startRowIndex_Equity <= separatorRowIndex ||
		startRowIndex_OtherEquity <= separatorRowIndex ||
		startRowIndex_CurrentAssets <= separatorRowIndex ||
		startRowIndex_CurrentLiabilities <= separatorRowIndex ||
		startRowIndex_NonCurrentAssets <= separatorRowIndex ||
		startRowIndex_NonCurrentLiabilities <= separatorRowIndex {
		err := errors.New("a Balance sheet section is above the separatorRowIndex")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	//if an inconsistency is found, return an error and ChatGPT api will be needed to process this R file
	if startRowIndex_Assets > endRowIndex_Assets ||
		startRowIndex_Liabilities > endRowIndex_Liabilities ||
		startRowIndex_Equity > endRowIndex_Equity ||
		startRowIndex_OtherEquity > endRowIndex_OtherEquity {
		err := errors.New("inconsistency in row indexes of balance sheet sections 1")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}
	if startRowIndex_CurrentAssets > endRowIndex_CurrentAssets ||
		startRowIndex_CurrentLiabilities > endRowIndex_CurrentLiabilities ||
		startRowIndex_NonCurrentAssets > endRowIndex_NonCurrentAssets ||
		startRowIndex_NonCurrentLiabilities > endRowIndex_NonCurrentLiabilities {
		err := errors.New("inconsistency in row indexes of balance sheet sections 2")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	//current is always above non current, asset is always above liability, liability is always above equity, Other equity is always below liability
	//Assets, liabilities, equity and other equity cannot overlap
	//current and non current cannot overlap

	// Current above non-current
	if endRowIndex_CurrentAssets > startRowIndex_NonCurrentAssets {
		err := errors.New("current assets section overlaps with non-current assets")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}
	if endRowIndex_CurrentLiabilities > startRowIndex_NonCurrentLiabilities {
		err := errors.New("current liabilities section overlaps with non-current liabilities")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	// Assets above liabilities above equity
	if endRowIndex_Assets > startRowIndex_Liabilities {
		err := errors.New("assets section overlaps with liabilities")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}
	if endRowIndex_Liabilities > startRowIndex_Equity {
		err := errors.New("liabilities section overlaps with equity")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	// Other equity below liability
	if startRowIndex_OtherEquity != 0 && startRowIndex_OtherEquity < endRowIndex_Liabilities {
		err := errors.New("other equity section appears before end of liabilities")
		fmt.Println(err)
		return BalanceSheetIndices{}, err
	}

	// Check for section overlaps
	if startRowIndex_OtherEquity != 0 {
		// Check if Other Equity's range (start to end) overlaps with any other section's range
		if (startRowIndex_OtherEquity <= endRowIndex_Assets && endRowIndex_OtherEquity >= startRowIndex_Assets) ||
			(startRowIndex_OtherEquity <= endRowIndex_Liabilities && endRowIndex_OtherEquity >= startRowIndex_Liabilities) ||
			(startRowIndex_OtherEquity <= endRowIndex_Equity && endRowIndex_OtherEquity >= startRowIndex_Equity) {
			err := errors.New("other equity section overlaps with other sections")
			fmt.Println(err)
			return BalanceSheetIndices{}, err
		}
	}

	//once the checks are done, return the indices struct
	indices := BalanceSheetIndices{
		ReportDate:                 reportDate,
		Form:                       form,
		AccessionNumber:            accessionNumber,
		TotalLineItemCount:         totalLineItemCount,
		SeparatorRowIndex:          separatorRowIndex,
		AssetsStart:                startRowIndex_Assets,
		AssetsEnd:                  endRowIndex_Assets,
		LiabilitiesStart:           startRowIndex_Liabilities,
		LiabilitiesEnd:             endRowIndex_Liabilities,
		EquityStart:                startRowIndex_Equity,
		EquityEnd:                  endRowIndex_Equity,
		OtherEquityStart:           startRowIndex_OtherEquity,
		OtherEquityEnd:             endRowIndex_OtherEquity,
		CurrentAssetsStart:         startRowIndex_CurrentAssets,
		CurrentAssetsEnd:           endRowIndex_CurrentAssets,
		NonCurrentAssetsStart:      startRowIndex_NonCurrentAssets,
		NonCurrentAssetsEnd:        endRowIndex_NonCurrentAssets,
		CurrentLiabilitiesStart:    startRowIndex_CurrentLiabilities,
		CurrentLiabilitiesEnd:      endRowIndex_CurrentLiabilities,
		NonCurrentLiabilitiesStart: startRowIndex_NonCurrentLiabilities,
		NonCurrentLiabilitiesEnd:   endRowIndex_NonCurrentLiabilities,
	}

	fmt.Println(indices)
	return indices, nil
}

// This function finds the common fields for all financial statement rfiles and returns them
// FirstEmptyRow serves as the separation row between header cells and data cells of R files that have been parsed
func CommonInitialProcessorForCsvRfile(RfilePath string) (RfileData2Darray_ [][]string, reportDate_ string, form_ string, accessionNumber_ string, totalLineItemCount_ int, separatorRowIndex_ int, err error) {
	RfileData2Darray, err := utilityfunctions.ReadCsvFileToArray(RfilePath)
	if err != nil {
		fmt.Println(err)
		return nil, "", "", "", 0, 0, err
	}

	if len(RfileData2Darray) == 0 {
		err := errors.New("empty CSV file")
		fmt.Println(err)
		return nil, "", "", "", 0, 0, err
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
		err := errors.New("reportDate not found")
		fmt.Println(err)
		return nil, "", "", "", 0, 0, err
	}
	if form == "" {
		err := errors.New("form not found")
		fmt.Println(err)
		return nil, "", "", "", 0, 0, err
	}
	if accessionNumber == "" {
		err := errors.New("accessionNumber not found")
		fmt.Println(err)
		return nil, "", "", "", 0, 0, err
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
