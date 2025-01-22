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

func CombineTwoBalanceSheets(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string) (CombinedBalanceSheet [][]string) {

	//go thru the combinedBalanceSheetLineItems and essentailly create a new balance sheet
	//for new balance sheet, we basically draw out the left col and the top rows for dates n stuff
	// and for each cell we do find a value that matches all the left col and top rows for the given cell in two input balancesheet arrays

	separatorRowIndex := -1
	accessionNumberRowIndex := -1
	formIndex := -1
	reportDateIndex := -1
	reportPeriodIndex := -1

	for rowIndex, row := range BalanceSheet1Array {
		switch row[0] {
		case "separator":
			separatorRowIndex = rowIndex
		case "accessionNumber":
			accessionNumberRowIndex = rowIndex
		case "form":
			formIndex = rowIndex
		case "reportDate":
			reportDateIndex = rowIndex
		case "reportPeriod":
			reportPeriodIndex = rowIndex
		default:
			// handle any other cases
		}
	}

	var combinedBalanceSheetArray [][]string
	//add in metadata of first balance sheet
	for i := 0; i < separatorRowIndex; i++ {
		combinedBalanceSheetArray = append(combinedBalanceSheetArray, BalanceSheet1Array[i])
	}
	//add in metadata of second balance sheet
	for i := 0; i < separatorRowIndex; i++ {
		for j := 1; j < len(BalanceSheet2Array[0]); j++ {
			combinedBalanceSheetArray[i] = append(combinedBalanceSheetArray[i], BalanceSheet2Array[i][j])
		}
	}
	//get the index of the cols in order it should be in
	rearrangedColumnIndices := GetIndexOfRearrangedColumnsByReportPeriodAndDate(combinedBalanceSheetArray, reportPeriodIndex, reportDateIndex)
	//rearrange the columns in order
	for targetIndex, sourceIndex := range rearrangedColumnIndices {
		if targetIndex+1 != sourceIndex {
			combinedBalanceSheetArray = RearrangeColumns(combinedBalanceSheetArray, sourceIndex, targetIndex+1)
		}
	}

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
	HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, CurrentAssetsLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Current Assets"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Non Current Assets"})
	HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, NonCurrentAssetsLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Assets"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Current Liabilities"})
	HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, CurrentLiabilitiesLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Current Liabilities"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Non Current Liabilities"})
	HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, NonCurrentLiabilitiesLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Liabilities"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Stockholders Equity"})
	HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, StockholdersEquityLineItemNames)
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Stockholders' Equity"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Total Liabilities and Stockholders' Equity"})
	combinedBalanceSheetArray = append(combinedBalanceSheetArray, []string{"Other Equities"})
	HelperFunction_AppendLineItemNamesToBalanceSheetArray(combinedBalanceSheetArray, OtherEquitiesLineItemNames)
}

func HelperFunction_AppendLineItemNamesToBalanceSheetArray(BalanceSheetArray [][]string, lineItemNames []string) {
	//add in line item names
	for i := 0; i < len(lineItemNames); i++ {
		BalanceSheetArray = append(BalanceSheetArray, []string{lineItemNames[i]})
	}
}

// this function is used to comnbine a section of two balance sheets line item names
// eg combine current assets
// the new line items will be added to the last row of the first balance sheet
func HelperFunction_CombineBalanceSheetSectionLineItemNames(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string, startIndex1 int, endIndex1 int, startIndex2 int, endIndex2 int) ([]string, error) {
	var combinedLineItemsNames []string

	// Process BalanceSheet1
	for i := startIndex1; i < endIndex1; i++ {
		row := BalanceSheet1Array[i]
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) {
			combinedLineItemsNames = append(combinedLineItemsNames, lineItemName)
		} else {
			err := fmt.Errorf("row %d is empty", i)
			fmt.Println(err)
			return nil, err
		}
	}

	// Process BalanceSheet2
	for i := startIndex2; i < endIndex2; i++ {
		row := BalanceSheet2Array[i]
		lineItemName := row[0]
		if DoesDataCellExistInThisRow(row) && !CheckIfLineItemNameIsInLineItemNameList(lineItemName, combinedLineItemsNames) {
			combinedLineItemsNames = append(combinedLineItemsNames, lineItemName)
		} else {
			err := fmt.Errorf("row %d is empty", i)
			fmt.Println(err)
			return nil, err
		}
	}

	return combinedLineItemsNames, nil
}

func CombineLineItemNamesOfTwoBalanceSheetsIntoOne(BalanceSheet1Array [][]string, BalanceSheet2Array [][]string, BalanceSheet1Classifications BalanceSheetLineItemClassifications, BalanceSheet2Classifications BalanceSheetLineItemClassifications) (CombinedBalanceSheetLineItems map[string][]string, err error) {
	// Combine current assets
	BalanceSheetCombinedCurrentAssetLineItems, err := HelperFunction_CombineBalanceSheetSectionLineItemNames(
		BalanceSheet1Array, BalanceSheet2Array,
		BalanceSheet1Classifications.CurrentAssets, BalanceSheet1Classifications.TotalAssets,
		BalanceSheet2Classifications.CurrentAssets, BalanceSheet2Classifications.TotalAssets,
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
		BalanceSheet1Classifications.CurrentLiabilities, BalanceSheet1Classifications.TotalLiabilities,
		BalanceSheet2Classifications.CurrentLiabilities, BalanceSheet2Classifications.TotalLiabilities,
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
		BalanceSheet1Classifications.StockholdersEquity, BalanceSheet1Classifications.TotalStockholdersEquity,
		BalanceSheet2Classifications.StockholdersEquity, BalanceSheet2Classifications.TotalStockholdersEquity,
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
	}, nil

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

	var accessionNumber string = BalanceSheetArray[0][1]
	var separatorRowIndex int
	for i, row := range BalanceSheetArray {
		if row[0] == "separator" {
			separatorRowIndex = i
			break
		}
	}

	//find all the line items in the balance sheet
	for i := separatorRowIndex + 1; i < len(BalanceSheetArray); i++ {
		row := BalanceSheetArray[i]

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
		lineItemName := BalanceSheetArray[rowIndex][0]
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
