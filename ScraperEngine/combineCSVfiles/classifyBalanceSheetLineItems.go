package combinecsvfiles

import (
	"fmt"

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

func TesterFunction(CIK string, client *mongo.Client) {
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
	// for i := 0; i < len(balanceSheetLineItemClassificationsSlice)-1; i++ {
	// 	combinedBalanceSheetArray = CombineTwoBalanceSheets(combinedBalanceSheetArray, BalanceSheetArrays[i+1])
	// }
	for i := 0; i < 1; i++ {
		combinedBalanceSheetArray = CombineTwoBalanceSheets(combinedBalanceSheetArray, BalanceSheetArrays[i+1])
	}
	// fmt.Printf("Combined balance sheet: %+v\n", combinedBalanceSheetArray)
}

func CombineTwoBalanceSheets(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string) (CombinedBalanceSheet [][]string) {

	//go thru the combinedBalanceSheetLineItems and essentailly create a new balance sheet
	//for new balance sheet, we basically draw out the left col and the top rows for dates n stuff
	// and for each cell we do find a value that matches all the left col and top rows for the given cell in two input balancesheet arrays

	separatorRowIndex := -1
	// accessionNumberRowIndex := -1
	// formIndex := -1
	reportDateIndex := -1
	reportPeriodIndex := -1

	for i, row := range BalanceSheet1Array {
		if len(row) == 0 {
			continue // Skip empty rows
		}
		if row[0] == "separator" {
			separatorRowIndex = i
			break
		}
		if row[0] == "reportDate" {
			reportDateIndex = i
		}
		if row[0] == "reportPeriod" {
			reportPeriodIndex = i
		}
		// if row[0] == "accessionNumber" {
		// 	accessionNumberRowIndex = i
		// }
		// if row[0] == "form" {
		// 	formIndex = i
		// }
	}

	var combinedBalanceSheetArray [][]string
	//add in metadata of first balance sheet
	for i := 0; i <= separatorRowIndex; i++ {
		combinedBalanceSheetArray = append(combinedBalanceSheetArray, BalanceSheet1Array[i])
	}
	//add in metadata of second balance sheet
	for i := 0; i <= separatorRowIndex; i++ {
		for j := 1; j < len(BalanceSheet2Array[0]); j++ {
			combinedBalanceSheetArray[i] = append(combinedBalanceSheetArray[i], BalanceSheet2Array[i][j])
		}
	}

	//get the index of the cols in order it should be in
	rearrangedColumnIndices := GetIndexOfRearrangedColumnsByReportPeriodAndDate(combinedBalanceSheetArray, reportPeriodIndex, reportDateIndex)

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
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Stockholders Equity"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, StockholdersEquityLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Stockholders' Equity"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Liabilities and Stockholders' Equity"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Other Equities"})
	combinedBalanceSheetArray = HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, OtherEquitiesLineItemNames)

	fmt.Println(combinedBalanceSheetArray)

	return combinedBalanceSheetArray
	//now do lookup and fill in the values
	//need to account for the fact that some line items in other equities have exact same so we need to keep track of which index have been accounted for in each balance sheet
	//basically once we fill in a cell, we need to keep track that particualr cell has been used in individual balance sheet
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
	var separatorRowIndex int
	for i, row := range BalanceSheetArray {
		if len(row) == 0 {
			continue // Skip empty rows
		}
		if row[0] == "separator" {
			separatorRowIndex = i
			break
		}
	}

	//find all the line items in the balance sheet
	for i := separatorRowIndex + 1; i < len(BalanceSheetArray); i++ {
		row := BalanceSheetArray[i]
		if len(row) == 0 {
			continue // Skip empty rows
		}
		if containsAny(row[0], []string{"Current Assets", "Current assets"}) && currentAssetsRowIndex == -1 {
			currentAssetsRowIndex = i
		}
		if containsAny(row[0], []string{"Total Current Assets", "Total current assets"}) && totalCurrentAssetsRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalCurrentAssetsRowIndex = i
		}
		if containsAny(row[0], []string{"Total Assets", "Total assets"}) && totalAssetsRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalAssetsRowIndex = i
		}
		if containsAny(row[0], []string{"Current Liabilities", "Current liabilities"}) && currentLiabilitiesRowIndex == -1 {
			currentLiabilitiesRowIndex = i
		}
		if containsAny(row[0], []string{"Total Current Liabilities", "Total current liabilities"}) && totalCurrentLiabilitiesRowIndex == -1 && DoesDataCellExistInThisRow(row) {
			totalCurrentLiabilitiesRowIndex = i
		}
		if containsAny(row[0], []string{"Total Liabilities", "Total liabilities"}) && totalLiabilitiesRowIndex == -1 && DoesDataCellExistInThisRow(row) {
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

	// Check the order using the helper function
	// Create a map of indices for order checking
	testIndices := map[string]int{
		"currentAssets":           currentAssetsRowIndex,
		"totalCurrentAssets":      totalCurrentAssetsRowIndex,
		"totalAssets":             totalAssetsRowIndex,
		"currentLiabilities":      currentLiabilitiesRowIndex,
		"totalCurrentLiabilities": totalCurrentLiabilitiesRowIndex,
		"totalLiabilities":        totalLiabilitiesRowIndex,
		"stockholdersEquity":      stockholdersEquityRowIndex,
		"totalStockholdersEquity": totalStockholdersEquityRowIndex,
		"totalLiabilitiesEquity":  totalLiabilitiesEquityAndOtherEquityRowIndex,
	}
	if err := CheckBalanceSheetOrder(testIndices); err != nil {
		return BalanceSheetLineItemClassifications{}, fmt.Errorf("balance sheet order error: %v", err)
	}

	//check if all data cell rows are accouneted for
	var dataCellRowIndex []int
	for i := separatorRowIndex + 1; i < len(BalanceSheetArray); i++ {
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
