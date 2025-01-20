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

// this function is used to comnbine a section of two balance sheets
// eg combine current assets
// the new line items will be added to the last row of the first balance sheet
func HelperFunction_CombineBalanceSheetSectionLineItemNames(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string, startIndex1 int, endIndex1 int, startIndex2 int, endIndex2 int) []string {
	var combinedLineItemsNames []string

	// Process BalanceSheet1
	for i := startIndex1; i < endIndex1; i++ {
		row := BalanceSheet1Array[i]
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) {
			combinedLineItemsNames = append(combinedLineItemsNames, lineItemName)
		}
	}

	// Process BalanceSheet2
	for i := startIndex2; i < endIndex2; i++ {
		row := BalanceSheet2Array[i]
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) && !CheckIfLineItemNameIsInLineItemNameList(lineItemName, combinedLineItemsNames) {
			combinedLineItemsNames = append(combinedLineItemsNames, lineItemName)
		}
	}

	return combinedLineItemsNames
}

func combineLineItemNamesOfTwoBalanceSheetsIntoOne(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string, BalanceSheet1Classifications BalanceSheetLineItemClassifications, BalanceSheet2Classifications BalanceSheetLineItemClassifications) map[string][]string {

	// Combine current assets
	BalanceSheetCombinedCurrentAssetLineItems := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.CurrentAssets, BalanceSheet1Classifications.TotalAssets,
		BalanceSheet2Classifications.CurrentAssets, BalanceSheet2Classifications.TotalAssets,
	)
	// Combine non-current assets
	BalanceSheetCombinedNonCurrentAssetLineItems := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.TotalCurrentAssets+1, BalanceSheet1Classifications.TotalAssets,
		BalanceSheet2Classifications.TotalCurrentAssets+1, BalanceSheet2Classifications.TotalAssets,
	)
	// Combine current liabilities
	BalanceSheetCombinedCurrentLiabilitiesLineItems := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.CurrentLiabilities, BalanceSheet1Classifications.TotalLiabilities,
		BalanceSheet2Classifications.CurrentLiabilities, BalanceSheet2Classifications.TotalLiabilities,
	)
	// Combine non-current liabilities
	BalanceSheetCombinedNonCurrentLiabilitiesLineItems := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.TotalCurrentLiabilities+1, BalanceSheet1Classifications.TotalLiabilities,
		BalanceSheet2Classifications.TotalCurrentLiabilities+1, BalanceSheet2Classifications.TotalLiabilities,
	)
	// Combine stockholders equity
	BalanceSheetCombinedStockholdersEquityLineItems := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.StockholdersEquity, BalanceSheet1Classifications.TotalStockholdersEquity,
		BalanceSheet2Classifications.StockholdersEquity, BalanceSheet2Classifications.TotalStockholdersEquity,
	)
	// Combine other equities
	BalanceSheetCombinedOtherEquitiesLineItems := []string{}
	//loop thru other equities indexes of BS1 and add it to combinedOtherEquitiesLineItems
	//loop thru other equities indexes of BS2 and compare
	for _, rowIndex := range BalanceSheet1Classifications.OtherEquities {
		row := BalanceSheet1Array[rowIndex]
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) {
			BalanceSheetCombinedOtherEquitiesLineItems = append(BalanceSheetCombinedOtherEquitiesLineItems, lineItemName)
		}
	}
	for _, rowIndex := range BalanceSheet2Classifications.OtherEquities {
		row := BalanceSheet2Array[rowIndex]
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
	}
}

func TesterFunction(CIK string, client *mongo.Client) {
	BalanceSheetArrays, _, _, _, err := GetCsvRfilesIntoArrayVariables(CIK, client)
	if err != nil {
		fmt.Println("Error getting CSV files:", err)
		return
	}

	var failedAccessionNumbers []string

	for _, BalanceSheetArray := range BalanceSheetArrays {
		accessionNumber := BalanceSheetArray[0][1]
		BalanceSheetLineItemClassifications, err := classifyBalanceSheetLineItems(BalanceSheetArray)
		if err != nil {
			fmt.Printf("Error classifying balance sheet line items for accession number %s: %v\n", accessionNumber, err)
			failedAccessionNumbers = append(failedAccessionNumbers, accessionNumber)
			continue
		}
		fmt.Printf("Successfully classified balance sheet for accession number: %s\n", accessionNumber)
		fmt.Printf("Classifications: %+v\n", BalanceSheetLineItemClassifications)
	}

	if len(failedAccessionNumbers) > 0 {
		fmt.Println("\nFailed to classify the following accession numbers:")
		for _, accNum := range failedAccessionNumbers {
			fmt.Printf("- %s\n", accNum)
		}
		fmt.Printf("Total failures: %d\n", len(failedAccessionNumbers))
	}

}

func classifyBalanceSheetLineItems(financialStatementArray [][]string) (BalanceSheetLineItemClassifications, error) {
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

	var accessionNumber string = financialStatementArray[0][1]
	var separatorRowIndex int
	for i, row := range financialStatementArray {
		if row[0] == "separator" {
			separatorRowIndex = i
			break
		}
	}

	//find all the line items in the balance sheet
	for i := separatorRowIndex + 1; i < len(financialStatementArray); i++ {
		row := financialStatementArray[i]

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
		for i := totalLiabilitiesRowIndex + 1; i < len(financialStatementArray); i++ {
			row := financialStatementArray[i]
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
	for i := separatorRowIndex + 1; i < len(financialStatementArray); i++ {
		if DoesDataCellExistInThisRow(financialStatementArray[i]) {
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
			fmt.Println(
				"currentAssets", currentAssetsRowIndex,
				"totalCurrentAssets", totalCurrentAssetsRowIndex,
				"totalAssets", totalAssetsRowIndex,
				"currentLiabilities", currentLiabilitiesRowIndex,
				"totalCurrentLiabilities", totalCurrentLiabilitiesRowIndex,
				"totalLiabilities", totalLiabilitiesRowIndex,
				"stockholdersEquity", stockholdersEquityRowIndex,
				"totalStockholdersEquity", totalStockholdersEquityRowIndex,
				"totalLiabilitiesEquityAndOtherEquity", totalLiabilitiesEquityAndOtherEquityRowIndex,
				"otherEquities", otherEquitiesRowIndex,
			)
			return BalanceSheetLineItemClassifications{},
				fmt.Errorf("accession number: %s - found unaccounted data cell at row %d", accessionNumber, dataRow)
		}
	}

	//check if OtherEquities have duplicate line item names
	var lineItemNames []string
	for _, rowIndex := range otherEquitiesRowIndex {
		lineItemName := financialStatementArray[rowIndex][0]
		// Check for exact match
		for _, existingName := range lineItemNames {
			if lineItemName == existingName {
				return BalanceSheetLineItemClassifications{}, fmt.Errorf("duplicate lineItemName found in OtherEquities: %s", lineItemName)
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
