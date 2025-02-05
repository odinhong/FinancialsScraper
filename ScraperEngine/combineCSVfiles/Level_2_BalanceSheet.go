package combinecsvfiles

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	utilityFunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
)

func TesterFunctionForLevel2BalanceSheets(CIK string) [][]string {
	base_directory := filepath.Join("SEC-files", "combinedFinancialStatements")
	fileName := CIK + "_combinedBalanceSheetLevel1.csv"
	filepath := filepath.Join(base_directory, fileName)
	Level1BalanceSheetArray, err := utilityFunctions.ReadCsvFileToArray(filepath)
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return nil
	}
	Level1BalanceSheetArray, _ = DeleteDuplicateBalanceSheetColumnsWithSameReportPeriodAndKeepTheMostRecentReportDateColumn(Level1BalanceSheetArray)
	utilityFunctions.Save2DarrayToCsvFile(Level1BalanceSheetArray, base_directory, CIK+"_test.csv")
	fmt.Println("check 1")

	BalanceSheetArrayInProcess, _ := DeleteEmptyLineItemBalanceSheetRows(Level1BalanceSheetArray)
	fmt.Println("check 2")

	utilityFunctions.Save2DarrayToCsvFile(BalanceSheetArrayInProcess, base_directory, CIK+"_combinedBalanceSheetLevel2.csv")
	fmt.Println("Saved Balance Sheet Array to CSV file")
	return BalanceSheetArrayInProcess
}

func DeleteDuplicateBalanceSheetColumnsWithSameReportPeriodAndKeepTheMostRecentReportDateColumn(Level1BalanceSheetArray [][]string) (ProcessedBalanceSheetArray [][]string, err error) {
	//If the column has same reportPeriod & reportDurationInMonths on the right delete it
	for i := 1; i < len(Level1BalanceSheetArray[AccessionNumberRowIndex])-1; i++ { //-1 to prevent out of bounds error
		reportDate_left := Level1BalanceSheetArray[ReportDateRowIndex][i]
		reportPeriod_left := Level1BalanceSheetArray[ReportPeriodRowIndex][i]
		reportDate_right := Level1BalanceSheetArray[ReportDateRowIndex][i+1]
		reportPeriod_right := Level1BalanceSheetArray[ReportPeriodRowIndex][i+1]

		// Convert YYYYMMDD format strings directly to integers
		reportDateLeft, err := strconv.Atoi(reportDate_left)
		if err != nil {
			fmt.Printf("Error converting date %s to integer: %v\n", reportDate_left, err)
			return nil, err
		}
		reportDateRight, err := strconv.Atoi(reportDate_right)
		if err != nil {
			fmt.Printf("Error converting date %s to integer: %v\n", reportDate_right, err)
			return nil, err
		}

		//delete the column with the same reportPeriod but older reportDate
		if reportPeriod_left == reportPeriod_right {
			if reportDateLeft < reportDateRight { // Keep the more recent date (higher number)
				Level1BalanceSheetArray = DeleteColumn(Level1BalanceSheetArray, i)
				i-- // Decrement i since we removed a column and need to recheck the new adjacent columns
			} else {
				Level1BalanceSheetArray = DeleteColumn(Level1BalanceSheetArray, i+1)
				i-- // Decrement i since we removed a column and need to recheck the new adjacent columns
			}
		}

	}

	return Level1BalanceSheetArray, nil
}

func DeleteEmptyLineItemBalanceSheetRows(BalanceSheetArray [][]string) (ProcessedBalanceSheetArray [][]string, err error) {
	for i := SeparatorRowIndex + 1; i < len(BalanceSheetArray); i++ {
		//remove empty row
		if len(BalanceSheetArray[i]) == 0 {
			BalanceSheetArray = DeleteRow(BalanceSheetArray, i)
			i--
			continue
		}

		//skip row if it matches any of the balance sheet category names (case and space insensitive)
		skipRow := false
		for _, categoryName := range BalanceSheetCategoryNames { //BalanceSheetCategoryNames is defined in constants.go
			if strings.EqualFold(strings.ReplaceAll(BalanceSheetArray[i][0], " ", ""), strings.ReplaceAll(categoryName, " ", "")) {
				skipRow = true
				break
			}
		}
		if skipRow {
			continue
		}
		//delete row if it doesn't contain any of the balance sheet category names and does not contain any data cells
		if !DoesDataCellExistInThisRow(BalanceSheetArray[i]) {
			BalanceSheetArray = DeleteRow(BalanceSheetArray, i)
			i--
			continue
		}
	}

	return BalanceSheetArray, nil
}

//need a function to combine common balance sheet items
//eg) account receivable, common stock, accumulated other comprehensive income, convertibgle preferred stock
//they have a pattern wehre its the words then "," a comma follows to add addtional info
//eg) Common stock, $0.000006 par value; 5,000 million Class A shares authorized, 1,970 million and 1,671 million shares issued and outstanding, including 6 million and 2 million outstanding shares subject to repurchase as of December 31, 2013 and December 31, 2012,
//so maybe i just need to find a comma and see if the words b4 comma is same

// "account receivable, net"
// "common stock,"
// func MergeCertainCommonLineItemRows(BalanceSheetArray [][]string) (ProcessedBalanceSheetArray [][]string, err error) {
// 	CommonLineItemNames := [][]string{
// 		[]string{"Account receivable, net"},
// 		[]string{"Common stock,"},
// 		[]string{"Accumulated other comprehensive income", "loss"},
// 	}

// 	//find all line items within the balance sheet category that start with the one of the CommonLineItemNames
// 	//turn the line item name into

// 	return BalanceSheetArray, nil
// }
