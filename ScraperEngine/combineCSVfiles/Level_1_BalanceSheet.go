package combinecsvfiles

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	utilityFunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"

	"go.mongodb.org/mongo-driver/mongo"
)

type BalanceSheetLineItemClassifications struct {
	CurrentAssets                        int
	TotalCurrentAssets                   int
	TotalAssets                          int
	CurrentLiabilities                   int
	TotalCurrentLiabilities              int
	TotalLiabilities                     int
	StockholdersEquity                   int
	TotalStockholdersEquity              int
	TotalLiabilitiesEquityAndOtherEquity int
	OtherEquities                        []int
}

func GenerateLevel1CombinedBalanceSheetsAndSaveAsCsvFileGivenCIK(CIK string, client *mongo.Client) {
	BalanceSheetArrays, _, _, _, err := GetCsvRfilesIntoArrayVariables(CIK, client)
	if err != nil {
		fmt.Println("Error getting CSV files:", err)
		return
	}

	var failedAccessionNumbers []string
	var balanceSheetLineItemClassificationsSlice []BalanceSheetLineItemClassifications
	for i := 0; i < len(BalanceSheetArrays); i++ {
		BalanceSheetArray := BalanceSheetArrays[i]
		accessionNumber := BalanceSheetArray[0][1]
		BalanceSheetLineItemClassifications, err := classifyBalanceSheetLineItems(BalanceSheetArray)
		if err != nil {
			fmt.Printf("Error classifying balance sheet line items for accession number %s: %v\n", accessionNumber, err)
			failedAccessionNumbers = append(failedAccessionNumbers, accessionNumber)
			//remove the balance sheet from the slice
			BalanceSheetArrays = append(BalanceSheetArrays[:i], BalanceSheetArrays[i+1:]...)
			i-- // Decrement i since we removed an element
			continue
		}
		balanceSheetLineItemClassificationsSlice = append(balanceSheetLineItemClassificationsSlice, BalanceSheetLineItemClassifications)
		// fmt.Printf("Successfully classified balance sheet for accession number: %s\n", accessionNumber)
		// fmt.Printf("Classifications: %+v\n", BalanceSheetLineItemClassifications)
	}
	if len(failedAccessionNumbers) > 0 {
		fmt.Println("\nFailed to classify the following accession numbers:")
		for _, accNum := range failedAccessionNumbers {
			fmt.Printf("- %s\n", accNum)
		}
		fmt.Printf("Total failures: %d\n", len(failedAccessionNumbers))
	}
	if len(balanceSheetLineItemClassificationsSlice) != len(BalanceSheetArrays) {
		fmt.Println("Error: Number of balance sheets classified does not match number of balance sheets")
	}
	// fmt.Println(balanceSheetLineItemClassificationsSlice)

	combinedBalanceSheetArray := BalanceSheetArrays[0]
	for i := 0; i < len(balanceSheetLineItemClassificationsSlice)-1; i++ {
		combinedBalanceSheetArray = CombineTwoBalanceSheets(combinedBalanceSheetArray, BalanceSheetArrays[i+1])
	}

	fmt.Print("Combined Balance Sheet Array: [\n")
	for i, row := range combinedBalanceSheetArray {
		fmt.Printf("  [%s]", strings.Join(row, ", "))
		if i < len(combinedBalanceSheetArray)-1 {
			fmt.Print(",")
		}
		fmt.Print("\n")
	}
	fmt.Print("]\n")

	//convert combinedBalanceSheetArray to csv file and save it to SEC-files/combinedFinancialStatements
	directory := filepath.Join("SEC-files", "combinedFinancialStatements")
	fileName := CIK + "_combinedBalanceSheetLevel1.csv"
	if err := utilityFunctions.Save2DarrayToCsvFile(combinedBalanceSheetArray, directory, fileName); err != nil {
		fmt.Printf("Error saving CSV file: %v\n", err)
		return
	}
	fmt.Printf("Successfully saved combined balance sheet to %s\n", filepath.Join(directory, fileName))

}

func CombineTwoBalanceSheets(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string) (CombinedBalanceSheet [][]string) {
	//go thru the combinedBalanceSheetLineItems and essentailly create a new balance sheet
	//for new balance sheet, we basically draw out the left col and the top rows for dates n stuff
	// and for each cell we do find a value that matches all the left col and top rows for the given cell in two input balancesheet arrays

	var combinedBalanceSheetArray [][]string
	//add in metadata of first balance sheet
	for i := 0; i <= SeparatorRowIndex; i++ {
		combinedBalanceSheetArray = append(combinedBalanceSheetArray, BalanceSheet1Array[i])
	}
	//add in metadata of second balance sheet
	for i := 0; i <= SeparatorRowIndex; i++ {
		for j := 1; j < len(BalanceSheet2Array[0]); j++ {
			combinedBalanceSheetArray[i] = append(combinedBalanceSheetArray[i], BalanceSheet2Array[i][j])
		}
	}

	//get the index of the cols in order it should be in
	rearrangedColumnIndices := GetIndexOfRearrangedColumnsByReportPeriodAndDate(combinedBalanceSheetArray, ReportPeriodRowIndex, ReportDateRowIndex)

	// Rearrange all columns at once
	combinedBalanceSheetArray = RearrangeAllColumns(combinedBalanceSheetArray, rearrangedColumnIndices)

	//add in line item names
	BalanceSheet1Classifications, err := classifyBalanceSheetLineItems(BalanceSheet1Array)
	if err != nil {
		fmt.Println(err)
	}
	BalanceSheet2Classifications, err := classifyBalanceSheetLineItems(BalanceSheet2Array)
	if err != nil {
		fmt.Println(err)
	}

	//convert the title line item names to the names I want to use, otherwise FillInDataCells wont work properly
	//ex) if the OG line item name is "Total Shareholders' Equity" but since the combinedBalanceSheet has "Total Stockholders' Equity" and FIllinDataCEll funciton only matches exact match so it won't find the line item.
	//resulting in a combinedbalance sheet with missing data cells
	// we do this for all the title line except for Other Equities
	BalanceSheet1Array[BalanceSheet1Classifications.CurrentAssets][0] = "Current Assets"
	BalanceSheet1Array[BalanceSheet1Classifications.TotalCurrentAssets][0] = "Total Current Assets"
	BalanceSheet1Array[BalanceSheet1Classifications.TotalAssets][0] = "Total Assets"
	BalanceSheet1Array[BalanceSheet1Classifications.CurrentLiabilities][0] = "Current Liabilities"
	BalanceSheet1Array[BalanceSheet1Classifications.TotalCurrentLiabilities][0] = "Total Current Liabilities"
	BalanceSheet1Array[BalanceSheet1Classifications.TotalLiabilities][0] = "Total Liabilities"
	BalanceSheet1Array[BalanceSheet1Classifications.StockholdersEquity][0] = "Stockholders' Equity"
	BalanceSheet1Array[BalanceSheet1Classifications.TotalStockholdersEquity][0] = "Total Stockholders' Equity"
	BalanceSheet1Array[BalanceSheet1Classifications.TotalLiabilitiesEquityAndOtherEquity][0] = "Total Liabilities and Stockholders' Equity"

	BalanceSheet2Array[BalanceSheet2Classifications.CurrentAssets][0] = "Current Assets"
	BalanceSheet2Array[BalanceSheet2Classifications.TotalCurrentAssets][0] = "Total Current Assets"
	BalanceSheet2Array[BalanceSheet2Classifications.TotalAssets][0] = "Total Assets"
	BalanceSheet2Array[BalanceSheet2Classifications.CurrentLiabilities][0] = "Current Liabilities"
	BalanceSheet2Array[BalanceSheet2Classifications.TotalCurrentLiabilities][0] = "Total Current Liabilities"
	BalanceSheet2Array[BalanceSheet2Classifications.TotalLiabilities][0] = "Total Liabilities"
	BalanceSheet2Array[BalanceSheet2Classifications.StockholdersEquity][0] = "Stockholders' Equity"
	BalanceSheet2Array[BalanceSheet2Classifications.TotalStockholdersEquity][0] = "Total Stockholders' Equity"
	BalanceSheet2Array[BalanceSheet2Classifications.TotalLiabilitiesEquityAndOtherEquity][0] = "Total Liabilities and Stockholders' Equity"

	combinedBalanceSheetLineItems, err := CombineLineItemNamesOfTwoBalanceSheetsIntoOne(BalanceSheet1Array, BalanceSheet2Array, BalanceSheet1Classifications, BalanceSheet2Classifications)
	if err != nil {
		fmt.Println(err)
	}
	CurrentAssetsLineItemNames := combinedBalanceSheetLineItems["CurrentAssetsLineItemNames"]
	NonCurrentAssetsLineItemNames := combinedBalanceSheetLineItems["NonCurrentAssetsLineItemNames"]
	CurrentLiabilitiesLineItemNames := combinedBalanceSheetLineItems["CurrentLiabilitiesLineItemNames"]
	NonCurrentLiabilitiesLineItemNames := combinedBalanceSheetLineItems["NonCurrentLiabilitiesLineItemNames"]
	StockholdersEquityLineItemNames := combinedBalanceSheetLineItems["StockholdersEquityLineItemNames"]
	OtherEquitiesLineItemNames := combinedBalanceSheetLineItems["OtherEquitiesLineItemNames"]

	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Current Assets"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, CurrentAssetsLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Current Assets"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Non Current Assets"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, NonCurrentAssetsLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Assets"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Current Liabilities"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, CurrentLiabilitiesLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Current Liabilities"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Non Current Liabilities"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, NonCurrentLiabilitiesLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Liabilities"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Stockholders' Equity"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, StockholdersEquityLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Stockholders' Equity"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Liabilities and Stockholders' Equity"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Other Equities"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, OtherEquitiesLineItemNames)

	//add in "" empty string cells for FillInDataCellsToEmptyCombinedStatement to fill in
	//using first row which is accessionNumber row to know how many cells to add
	targetLength := len(combinedBalanceSheetArray[0])
	for i := 0; i < len(combinedBalanceSheetArray); i++ {
		emptyStringsNeeded := targetLength - len(combinedBalanceSheetArray[i])
		if emptyStringsNeeded > 0 {
			emptyStrings := make([]string, emptyStringsNeeded)
			combinedBalanceSheetArray[i] = append(combinedBalanceSheetArray[i], emptyStrings...)
		}
	}

	combinedBalanceSheetArray, err = FillInDataCellsToEmptyCombinedStatement(combinedBalanceSheetArray, BalanceSheet1Array, BalanceSheet2Array)
	if err != nil {
		fmt.Println(err)
	}

	return combinedBalanceSheetArray
}

func HelperFunction_AppendLineItemNamesToBalanceSheetArray(BalanceSheetArray [][]string, lineItemNames []string) [][]string {
	//add in line item names
	updatedBalanceSheetArray := BalanceSheetArray
	for i := 0; i < len(lineItemNames); i++ {
		updatedBalanceSheetArray = append(updatedBalanceSheetArray, []string{lineItemNames[i]})
	}
	return updatedBalanceSheetArray
}

// this function is used to comnbine a section of two balance sheets line item names
// eg combine current assets
// the new line items will be added to the last row of the first balance sheet
func HelperFunction_CombineBalanceSheetSectionLineItemNames(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string, startIndex1 int, endIndex1 int, startIndex2 int, endIndex2 int) ([]string, error) {
	var combinedLineItemsNames []string

	// Process BalanceSheet1
	for i := startIndex1; i < endIndex1; i++ {
		row := BalanceSheet1Array[i]
		if len(row) == 0 {
			continue // Skip empty rows
		}

		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) {
			combinedLineItemsNames = append(combinedLineItemsNames, lineItemName)
		} else {
			continue
		}
	}

	// Process BalanceSheet2
	for i := startIndex2; i < endIndex2; i++ {
		row := BalanceSheet2Array[i]
		if len(row) == 0 {
			continue // Skip empty rows
		}
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) && !CheckIfLineItemNameIsInLineItemNameList(lineItemName, combinedLineItemsNames) {
			combinedLineItemsNames = append(combinedLineItemsNames, lineItemName)
		} else {

			continue

		}
	}

	return combinedLineItemsNames, nil
}

func CombineLineItemNamesOfTwoBalanceSheetsIntoOne(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string, BalanceSheet1Classifications BalanceSheetLineItemClassifications, BalanceSheet2Classifications BalanceSheetLineItemClassifications) (CombinedBalanceSheetLineItems map[string][]string, err error) {
	// Combine current assets
	BalanceSheetCombinedCurrentAssetLineItems, err := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.CurrentAssets+1, BalanceSheet1Classifications.TotalCurrentAssets,
		BalanceSheet2Classifications.CurrentAssets+1, BalanceSheet2Classifications.TotalCurrentAssets,
		//adding one because the startingIndex is the row that contains header "current assets"
	)
	if err != nil {
		fmt.Println(err)
	}
	// Combine non-current assets
	BalanceSheetCombinedNonCurrentAssetLineItems, err := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.TotalCurrentAssets+1, BalanceSheet1Classifications.TotalAssets,
		BalanceSheet2Classifications.TotalCurrentAssets+1, BalanceSheet2Classifications.TotalAssets,
	)
	if err != nil {
		fmt.Println(err)
	}
	// Combine current liabilities
	BalanceSheetCombinedCurrentLiabilitiesLineItems, err := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.CurrentLiabilities+1, BalanceSheet1Classifications.TotalCurrentLiabilities,
		BalanceSheet2Classifications.CurrentLiabilities+1, BalanceSheet2Classifications.TotalCurrentLiabilities,
		//adding one because the startingIndex is the row that contains header "current liabilities"
	)
	if err != nil {
		fmt.Println(err)
	}
	// Combine non-current liabilities
	BalanceSheetCombinedNonCurrentLiabilitiesLineItems, err := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.TotalCurrentLiabilities+1, BalanceSheet1Classifications.TotalLiabilities,
		BalanceSheet2Classifications.TotalCurrentLiabilities+1, BalanceSheet2Classifications.TotalLiabilities,
	)
	if err != nil {
		fmt.Println(err)
	}
	// Combine stockholders equity
	BalanceSheetCombinedStockholdersEquityLineItems, err := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.StockholdersEquity+1, BalanceSheet1Classifications.TotalStockholdersEquity,
		BalanceSheet2Classifications.StockholdersEquity+1, BalanceSheet2Classifications.TotalStockholdersEquity,
		//adding one because the startingIndex is the row that contains header "stockholders equity"

	)
	if err != nil {
		fmt.Println(err)
	}
	// Combine other equities
	BalanceSheetCombinedOtherEquitiesLineItems := []string{}
	//loop thru other equities indexes of BS1 and add it to combinedOtherEquitiesLineItems
	//loop thru other equities indexes of BS2 and compare
	for _, rowIndex := range BalanceSheet1Classifications.OtherEquities {
		row := BalanceSheet1Array[rowIndex]
		if len(row) == 0 {
			continue // Skip empty rows
		}
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) {
			BalanceSheetCombinedOtherEquitiesLineItems = append(BalanceSheetCombinedOtherEquitiesLineItems, lineItemName)
		}
	}
	for _, rowIndex := range BalanceSheet2Classifications.OtherEquities {
		row := BalanceSheet2Array[rowIndex]
		if len(row) == 0 {
			continue // Skip empty rows
		}
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) && !CheckIfLineItemNameIsInLineItemNameList(lineItemName, BalanceSheetCombinedOtherEquitiesLineItems) {
			BalanceSheetCombinedOtherEquitiesLineItems = append(BalanceSheetCombinedOtherEquitiesLineItems, lineItemName)
		}
	}

	return map[string][]string{
		"CurrentAssetsLineItemNames":         BalanceSheetCombinedCurrentAssetLineItems,
		"NonCurrentAssetsLineItemNames":      BalanceSheetCombinedNonCurrentAssetLineItems,
		"CurrentLiabilitiesLineItemNames":    BalanceSheetCombinedCurrentLiabilitiesLineItems,
		"NonCurrentLiabilitiesLineItemNames": BalanceSheetCombinedNonCurrentLiabilitiesLineItems,
		"StockholdersEquityLineItemNames":    BalanceSheetCombinedStockholdersEquityLineItems,
		"OtherEquitiesLineItemNames":         BalanceSheetCombinedOtherEquitiesLineItems,
	}, nil

}

func classifyBalanceSheetLineItems(BalanceSheetArray [][]string) (BalanceSheetLineItemClassifications, error) {
	var (
		currentAssetsRowIndex                        int = -1 // Using -1 as sentinel value
		totalCurrentAssetsRowIndex                   int = -1
		totalAssetsRowIndex                          int = -1
		currentLiabilitiesRowIndex                   int = -1
		totalCurrentLiabilitiesRowIndex              int = -1
		totalLiabilitiesRowIndex                     int = -1
		stockholdersEquityRowIndex                   int = -1
		totalStockholdersEquityRowIndex              int = -1
		totalLiabilitiesEquityAndOtherEquityRowIndex int = -1
	)

	var otherEquitiesRowIndex []int

	// Check if array is empty or first row doesn't have enough elements
	if len(BalanceSheetArray) == 0 {
		return BalanceSheetLineItemClassifications{}, fmt.Errorf("empty balance sheet array")
	}
	if len(BalanceSheetArray[0]) < 2 {
		return BalanceSheetLineItemClassifications{}, fmt.Errorf("first row of balance sheet doesn't have enough elements: got %d, want at least 2", len(BalanceSheetArray[0]))
	}

	var accessionNumber string = BalanceSheetArray[0][1]

	//find all the line items in the balance sheet
	for i := SeparatorRowIndex + 1; i < len(BalanceSheetArray); i++ {
		row := BalanceSheetArray[i]
		if len(row) == 0 {
			continue // Skip empty rows
		}
		if containsAny(row[0], []string{"Current Assets"}) && currentAssetsRowIndex == -1 {
			currentAssetsRowIndex = i
		}
		if containsAny(row[0], []string{"Total Current Assets"}) && totalCurrentAssetsRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalCurrentAssetsRowIndex = i
		}
		if containsAny(row[0], []string{"Total Assets"}) && totalAssetsRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalAssetsRowIndex = i
		}
		if containsAny(row[0], []string{"Current Liabilities"}) && currentLiabilitiesRowIndex == -1 {
			currentLiabilitiesRowIndex = i
		}
		if containsAny(row[0], []string{"Total Current Liabilities"}) && totalCurrentLiabilitiesRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalCurrentLiabilitiesRowIndex = i
		}
		if containsAny(row[0], []string{"Total Liabilities"}) && totalLiabilitiesRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalLiabilitiesRowIndex = i
		}
		if containsAny(row[0], []string{"Stockholders' Equity", "Shareholders' Equity", "Shareowners' Equity", "Equity"}) && notContainsAny(row[0], []string{"Investment", "marketable", "securities"}) && stockholdersEquityRowIndex == -1 {
			stockholdersEquityRowIndex = i
		}
		if containsAny(row[0], []string{"Total Stockholders' Equity", "Total Shareholders' Equity", "Total Shareowners' Equity", "Total Equity"}) && notContainsAny(row[0], []string{"Investment", "marketable", "securities"}) && totalStockholdersEquityRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalStockholdersEquityRowIndex = i
		}
		if CheckWordsInOrder(row[0], []string{"Total Liabilities", "and", "Equity"}) && totalLiabilitiesEquityAndOtherEquityRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalLiabilitiesEquityAndOtherEquityRowIndex = i
		}

	}

	if totalLiabilitiesRowIndex != -1 && stockholdersEquityRowIndex != -1 && totalStockholdersEquityRowIndex != -1 && totalLiabilitiesEquityAndOtherEquityRowIndex != -1 {
		for i := totalLiabilitiesRowIndex + 1; i < len(BalanceSheetArray); i++ {
			row := BalanceSheetArray[i]
			if len(row) == 0 {
				continue // Skip empty rows
			}
			//if there are data cells after total liabilities, and before stockholders equity, then those are Other equity line items
			if i > totalLiabilitiesRowIndex &&
				i < stockholdersEquityRowIndex &&
				DoesDataCellExistInThisRow(row) {

				otherEquitiesRowIndex = append(otherEquitiesRowIndex, i)
			}
			//if there are data cells after total liabilities equity and other equity, then those are other equity line items
			if totalLiabilitiesEquityAndOtherEquityRowIndex < i &&
				DoesDataCellExistInThisRow(row) {

				otherEquitiesRowIndex = append(otherEquitiesRowIndex, i)
			}
		}
	}

	//return error if any of the variables are still -1
	if currentAssetsRowIndex == -1 ||
		totalCurrentAssetsRowIndex == -1 ||
		totalAssetsRowIndex == -1 ||
		currentLiabilitiesRowIndex == -1 ||
		totalCurrentLiabilitiesRowIndex == -1 ||
		totalLiabilitiesRowIndex == -1 ||
		stockholdersEquityRowIndex == -1 ||
		totalStockholdersEquityRowIndex == -1 ||
		totalLiabilitiesEquityAndOtherEquityRowIndex == -1 {

		return BalanceSheetLineItemClassifications{}, fmt.Errorf("accession number: %s - could not find all the line items in the balance sheet\n"+
			"currentAssetsRowIndex: %d\n"+
			"totalCurrentAssetsRowIndex: %d\n"+
			"totalAssetsRowIndex: %d\n"+
			"currentLiabilitiesRowIndex: %d\n"+
			"totalCurrentLiabilitiesRowIndex: %d\n"+
			"totalLiabilitiesRowIndex: %d\n"+
			"stockholdersEquityRowIndex: %d\n"+
			"totalStockholdersEquityRowIndex: %d\n"+
			"totalLiabilitiesEquityAndOtherEquityRowIndex: %d",
			accessionNumber,
			currentAssetsRowIndex,
			totalCurrentAssetsRowIndex,
			totalAssetsRowIndex,
			currentLiabilitiesRowIndex,
			totalCurrentLiabilitiesRowIndex,
			totalLiabilitiesRowIndex,
			stockholdersEquityRowIndex,
			totalStockholdersEquityRowIndex,
			totalLiabilitiesEquityAndOtherEquityRowIndex)
	}

	//check if all data cell rows are accouneted for
	var dataCellRowIndex []int
	for i := SeparatorRowIndex + 1; i < len(BalanceSheetArray); i++ {
		if len(BalanceSheetArray[i]) == 0 {
			continue // Skip empty rows
		}
		if DoesDataCellExistInThisRow(BalanceSheetArray[i]) {
			dataCellRowIndex = append(dataCellRowIndex, i)
		}
	}
	var accountedForRowIndex []int
	//currentAssetsRowIndex to totalAssetsRowIndex
	//currentLiabilitiesRowIndex to totalLiabilitiesRowIndex
	//stocholdersEquityRowIndex to totalLiabilitiesEquityAndOtherEquityRowIndex
	//otherEquitiesRowIndex
	for i := currentAssetsRowIndex + 1; i <= totalAssetsRowIndex; i++ {
		accountedForRowIndex = append(accountedForRowIndex, i)
	}
	for i := currentLiabilitiesRowIndex + 1; i <= totalLiabilitiesRowIndex; i++ {
		accountedForRowIndex = append(accountedForRowIndex, i)
	}
	for i := stockholdersEquityRowIndex + 1; i <= totalLiabilitiesEquityAndOtherEquityRowIndex; i++ {
		accountedForRowIndex = append(accountedForRowIndex, i)
	}
	accountedForRowIndex = append(accountedForRowIndex, otherEquitiesRowIndex...)

	// Check if there are any data cell rows that are not accounted for
	for _, dataRow := range dataCellRowIndex {
		found := false
		for _, accountedRow := range accountedForRowIndex {
			if dataRow == accountedRow {
				found = true
				break
			}
		}
		if !found {
			// fmt.Println(
			// 	"currentAssets", currentAssetsRowIndex,
			// 	"totalCurrentAssets", totalCurrentAssetsRowIndex,
			// 	"totalAssets", totalAssetsRowIndex,
			// 	"currentLiabilities", currentLiabilitiesRowIndex,
			// 	"totalCurrentLiabilities", totalCurrentLiabilitiesRowIndex,
			// 	"totalLiabilities", totalLiabilitiesRowIndex,
			// 	"stockholdersEquity", stockholdersEquityRowIndex,
			// 	"totalStockholdersEquity", totalStockholdersEquityRowIndex,
			// 	"totalLiabilitiesEquityAndOtherEquity", totalLiabilitiesEquityAndOtherEquityRowIndex,
			// 	"otherEquities", otherEquitiesRowIndex,
			// )
			return BalanceSheetLineItemClassifications{},
				fmt.Errorf("accession number: %s - found unaccounted data cell at row %d", accessionNumber, dataRow)
		}
	}

	//check if OtherEquities have duplicate line item names
	var lineItemNames []string
	for _, rowIndex := range otherEquitiesRowIndex {
		row := BalanceSheetArray[rowIndex]
		if len(row) == 0 {
			continue // Skip empty rows
		}
		lineItemName := row[0]
		// Check for exact match
		for _, existingName := range lineItemNames {
			if lineItemName == existingName {
				return BalanceSheetLineItemClassifications{}, fmt.Errorf("duplicate lineItemName found in the financial statement: OtherEquities: %s", lineItemName)
			}
		}
		lineItemNames = append(lineItemNames, lineItemName)
	}

	return BalanceSheetLineItemClassifications{
		CurrentAssets:                        currentAssetsRowIndex,
		TotalCurrentAssets:                   totalCurrentAssetsRowIndex,
		TotalAssets:                          totalAssetsRowIndex,
		CurrentLiabilities:                   currentLiabilitiesRowIndex,
		TotalCurrentLiabilities:              totalCurrentLiabilitiesRowIndex,
		TotalLiabilities:                     totalLiabilitiesRowIndex,
		StockholdersEquity:                   stockholdersEquityRowIndex,
		TotalStockholdersEquity:              totalStockholdersEquityRowIndex,
		TotalLiabilitiesEquityAndOtherEquity: totalLiabilitiesEquityAndOtherEquityRowIndex,
		OtherEquities:                        otherEquitiesRowIndex,
	}, nil
}

// PrintBalanceSheetFields demonstrates how to iterate through the fields of BalanceSheetLineItemClassifications
func PrintBalanceSheetFields(bs BalanceSheetLineItemClassifications) {
	// Get the reflect.Value of the struct
	v := reflect.ValueOf(bs)

	// Get the reflect.Type of the struct
	t := v.Type()

	// Iterate through all fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		// Handle the OtherEquities slice separately
		if fieldName == "OtherEquities" {
			fmt.Printf("Field: %s (slice), Value: %v\n", fieldName, field.Interface())
			// You can iterate through the slice if needed
			for j := 0; j < field.Len(); j++ {
				fmt.Printf("\tItem %d: %v\n", j, field.Index(j).Interface())
			}
		} else {
			fmt.Printf("Field: %s, Value: %v\n", fieldName, field.Interface())
		}
	}
}
