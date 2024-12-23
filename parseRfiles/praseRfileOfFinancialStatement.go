package parserfiles

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	utilityfunctions "github.com/Programmerdin/FinancialDataSite_Go/utilityFunctions"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xmlquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StatementData struct {
	Headers [][]string
	Data    [][]string
}

func ParseManyRfilesAndSaveAsCSVs(CIK string, client *mongo.Client) {
	accesionNumbers, Rfilenames, err := RetrieveRfileNamesAndAccessionNumbersFromMongoDB(CIK, client)
	if err != nil {
		fmt.Println("Error RetrieveRfileNamesAndAccessionNumbersFromMongoDB function:", err)
		return
	}

	var accessionNumbers_to_parse []string
	var Rfilenames_to_parse []string
	for i := 0; i < len(accesionNumbers); i++ {
		RfileName_CSV := Rfilenames[i]
		ext := filepath.Ext(Rfilenames[i])
		if ext == ".htm" || ext == ".html" || ext == ".xml" {
			RfileName_CSV = strings.TrimSuffix(Rfilenames[i], ext) + ".csv"
		}
		filePath := filepath.Join("SEC-files", "filingSummaryAndRfiles", CIK, accesionNumbers[i], RfileName_CSV)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			accessionNumbers_to_parse = append(accessionNumbers_to_parse, accesionNumbers[i])
			Rfilenames_to_parse = append(Rfilenames_to_parse, Rfilenames[i])
		}
	}

	for i := 0; i < len(accessionNumbers_to_parse); i++ {
		ParseRfileAndSaveAsCSV(CIK, accessionNumbers_to_parse[i], Rfilenames_to_parse[i], client)
	}
}

func ParseRfileAndSaveAsCSV(CIK, accessionNumber, RfileName string, client *mongo.Client) error {
	//check if RfileName is .htm or .html or .xml
	fileExt := filepath.Ext(RfileName)

	parsedRfile := StatementData{
		Headers: [][]string{},
		Data:    [][]string{},
	}
	var err error
	switch fileExt {
	case ".htm", ".html":
		parsedRfile, err = ParseHtmRfile(CIK, accessionNumber, RfileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
	case ".xml":
		parsedRfile, err = ParseXmlRfile(CIK, accessionNumber, RfileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	cleanParsedRfile := CleanParsedRfile(&parsedRfile, accessionNumber, client)

	err = saveParsedRfileAsCSV(&cleanParsedRfile, CIK, accessionNumber, RfileName)
	if err != nil {
		log.Fatalf("Failed to save CSV: %v", err)
		return err
	}

	return nil
}

func ParseHtmRfile(CIK, accessionNumber, RfileName string) (StatementData, error) {
	RfilePath := filepath.Join("SEC-files", "filingSummaryAndRfiles", CIK, accessionNumber, RfileName)
	file, err := os.Open(RfilePath)
	if err != nil {
		log.Fatalf("could not open file %s: %v", RfilePath, err)
	}
	defer file.Close()

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatalf("could not create goquery document: %v", err)
	}

	statementData := StatementData{
		Headers: [][]string{},
		Data:    [][]string{},
	}
	rowspanGreaterThan1Locations := [][2]int{}

	// Parse rows
	doc.Find("table.report > tbody > tr:not(:has(tr)), table.report > tr:not(:has(tr))").Each(func(index int, element *goquery.Selection) {
		var rowData []string
		cols := element.Find("td")
		ths := element.Find("th")

		isHeaderRow := ths.Length() > 0
		if isHeaderRow {
			ths.Each(func(i int, el *goquery.Selection) {
				text := el.Text()
				text = strings.TrimSpace(text)
				colspan, _ := strconv.Atoi(el.AttrOr("colspan", "1"))
				rowspan, _ := strconv.Atoi(el.AttrOr("rowspan", "1"))

				for j := 0; j < colspan; j++ {
					if j == 0 {
						rowData = append(rowData, text)
					} else {
						rowData = append(rowData, "")
					}
				}

				if rowspan > 1 {
					for k := 0; k < colspan; k++ {
						rowspanGreaterThan1Locations = append(rowspanGreaterThan1Locations, [2]int{index, i + k})
					}
				}
			})
			statementData.Headers = append(statementData.Headers, rowData)
		} else {
			cols.Each(func(i int, el *goquery.Selection) {
				text := el.Text()
				text = strings.TrimSpace(text)
				colspan, _ := strconv.Atoi(el.AttrOr("colspan", "1"))

				for j := 0; j < colspan; j++ {
					if j == 0 {
						rowData = append(rowData, text)
					} else {
						rowData = append(rowData, "")
					}
				}
			})
			statementData.Data = append(statementData.Data, rowData)
		}
	})

	// Add empty cells to deal with rowspan > 1
	for _, location := range rowspanGreaterThan1Locations {
		rowIndex := location[0] + 1
		columnIndex := location[1]
		if rowIndex < len(statementData.Headers) {
			// Make a slice that is a copy of the current row up to the columnIndex
			rowBefore := make([]string, len(statementData.Headers[rowIndex][:columnIndex]))
			copy(rowBefore, statementData.Headers[rowIndex][:columnIndex])

			// Append an empty string where we need to add an empty cell
			rowBefore = append(rowBefore, "")

			// Append the rest of the row after the columnIndex
			rowAfter := statementData.Headers[rowIndex][columnIndex:]

			// Combine the row before and after slices
			statementData.Headers[rowIndex] = append(rowBefore, rowAfter...)
		}
	}

	// Remove rows that only have empty string
	filterRows := func(rows [][]string) [][]string {
		filtered := [][]string{}
		for _, row := range rows {
			notEmpty := false
			for _, cell := range row {
				if strings.TrimSpace(cell) != "" {
					notEmpty = true
					break
				}
			}
			if notEmpty {
				filtered = append(filtered, row)
			}
		}
		return filtered
	}
	statementData.Headers = filterRows(statementData.Headers)
	statementData.Data = filterRows(statementData.Data)

	// Check if every row is the same as total_column_count
	checkRowLengths := func(rows [][]string) bool {
		if len(rows) == 0 {
			return true
		}
		length := len(rows[0])
		for _, row := range rows {
			if len(row) != length {
				return false
			}
		}
		return true
	}

	headersAllSameLength := checkRowLengths(statementData.Headers)
	dataAllSameLength := checkRowLengths(statementData.Data)

	isEveryRowLengthSame := headersAllSameLength && dataAllSameLength
	if !isEveryRowLengthSame {
		fmt.Printf("CIK: %s, AccessionNumber: %s, RfileName: %s\n", CIK, accessionNumber, RfileName)
		log.Fatal("Error: Not all rows have the same length, problem during parsing HTM")
	}

	return statementData, nil

}

func ParseXmlRfile(CIK, accessionNumber, RfileName string) (StatementData, error) {
	statementData := StatementData{
		Headers: [][]string{},
		Data:    [][]string{},
	}
	filePath := filepath.Join("SEC-files/filingSummaryAndRfiles", CIK, accessionNumber, RfileName)
	xmlBytes, err := os.ReadFile(filePath)
	if err != nil {
		// handle error
		return statementData, err
	}

	doc, err := xmlquery.Parse(bytes.NewReader(xmlBytes))
	if err != nil {
		// handle error
		return statementData, err
	}

	//Set up the correct number of inner slices needed for headers
	// XPath query to find all Label elements within the first Labels element
	firstLabelsNode := xmlquery.FindOne(doc, "//Labels[1]")
	firstLabelslabelNodes := xmlquery.Find(firstLabelsNode, "./Label")

	for i := 0; i < len(firstLabelslabelNodes); i++ {
		if i == 0 {
			ReportName := xmlquery.FindOne(doc, "//ReportName")
			if ReportName == nil { //use ReportLongName if ReportName doesn't exist
				ReportName = xmlquery.FindOne(doc, "//ReportLongName")
			}
			statementData.Headers = append(statementData.Headers, []string{ReportName.InnerText()}) //empty string for first element to account for the fact that dates don't have a line item name
		} else {
			statementData.Headers = append(statementData.Headers, []string{""}) //empty string for first element to account for the fact that dates don't have a line item name
		}
	}

	// XPath query to find all Label elements within Labels
	//This is grabbing dates and the 6 months ended
	labelNodes := xmlquery.Find(doc, "//Labels/Label")
	for _, labelNode := range labelNodes {
		// Get the value of the 'Label' attribute
		labelText := labelNode.SelectAttr("Label")
		labelId := labelNode.SelectAttr("Id")

		if labelId == "1" {
			statementData.Headers[0] = append(statementData.Headers[0], labelText)
		}
		if labelId == "2" {
			statementData.Headers[1] = append(statementData.Headers[1], labelText)
		}
	}

	//Set up the correct number of inner slices needed for statementData.Data
	RowNodes := xmlquery.Find(doc, "//Rows/Row")
	rowCount := len(RowNodes)
	for i := 0; i < rowCount; i++ {
		statementData.Data = append(statementData.Data, []string{})
	}
	//Add line item names
	RowLabelNodes := xmlquery.Find(doc, "//Rows/Row/Label")
	for i := 0; i < len(RowLabelNodes); i++ {
		statementData.Data[i] = append(statementData.Data[i], RowLabelNodes[i].InnerText())
	}

	//How many cells per row excluding name of the line item (aka Label)
	CellNodes := xmlquery.Find(doc, "//Rows/Row[2]/Cells/Cell")
	cellCount := len(CellNodes)
	//Get NumericAmount of each Cell
	NumericAmountNodes := xmlquery.Find(doc, "//Rows/Row/Cells/Cell/NumericAmount")
	//Add cell data next to line item name
	for i := 0; i < rowCount; i++ {
		for j := 0; j < cellCount; j++ {
			statementData.Data[i] = append(statementData.Data[i], NumericAmountNodes[i*cellCount+j].InnerText())
		}
	}

	// Deal with first Data row that sometimes is the name of the financial statement
	FirstDataRowElement := statementData.Data[0][0]
	FirstDataRowElementClean := strings.ToLower(strings.ReplaceAll(FirstDataRowElement, " ", ""))
	ReportNametest := statementData.Headers[0][0]
	ReportNameClean := strings.ToLower(strings.ReplaceAll(ReportNametest, " ", ""))
	check := strings.Contains(ReportNameClean, FirstDataRowElementClean)
	if check {
		//remove first row of statementData.Data
		statementData.Data = statementData.Data[1:]
	}

	//Deal with CURRENT ASSET and CURRENT LIABILITY rows
	for i := len(statementData.Data) - 1; i >= 0; i-- {
		// Assume that the row is only made up of "0" or "" apart from the first element
		rowIsOnly0OrEmpty := true

		// Start from the second element (index 1)
		for j := 1; j < cellCount+1; j++ {
			cell := statementData.Data[i][j]
			if cell != "0" && cell != "" {
				// If a cell is not "0" or "", set the flag to false and break
				rowIsOnly0OrEmpty = false
				break
			}
		}

		// If the flag is still true, the row is only made up of "0" or "" apart from the first element
		if rowIsOnly0OrEmpty {
			//replace row with empty string except for first element
			for j := 1; j < cellCount+1; j++ {
				statementData.Data[i][j] = ""
			}
		}
	}

	// Remove rows that are abstract rows in statementData.Data
	for i := len(statementData.Data) - 1; i >= 0; i-- {
		isLineItemAbstract := strings.Contains(statementData.Data[i][0], "Abstract")
		isLineItem0OrEmpty := statementData.Data[i][1] == "0" || statementData.Data[i][1] == ""
		if isLineItemAbstract && isLineItem0OrEmpty {
			statementData.Data = append(statementData.Data[:i], statementData.Data[i+1:]...)
		}
	}

	return statementData, nil
}

func CleanParsedRfile(statementData *StatementData, accessionNumber string, client *mongo.Client) StatementData {
	var statementDataArray [][]string

	// Convert headers and data to one single slice
	statementDataArray = append(statementDataArray, statementData.Headers...)
	statementDataArray = append(statementDataArray, statementData.Data...)

	// Go through every cell and replace cells that contain unwanted characters with empty strings
	unwantedCharacters := []string{"[", "]"}
	for i := range statementDataArray {
		for j := range statementDataArray[i] {
			for _, unwantedChar := range unwantedCharacters {
				if strings.Contains(statementDataArray[i][j], unwantedChar) {
					statementDataArray[i][j] = ""
				}
			}
		}
	}

	//make a slice of index numbers of cols of statementDataArray
	colIndexShortlist := make([]int, len(statementDataArray[0]))
	for i := range colIndexShortlist {
		colIndexShortlist[i] = i
	}

	//Remove col from colIndexShortlist if col is not empty
	for i := range statementDataArray {
		for _, value := range colIndexShortlist {
			if statementDataArray[i][value] != "" {
				//find value in colIndexShortlist slice and remove it
				ColToRemove := indexOf(colIndexShortlist, value)
				if ColToRemove != -1 {
					colIndexShortlist = append(colIndexShortlist[:ColToRemove], colIndexShortlist[ColToRemove+1:]...)
				}
			}
		}
	}

	//Remove the col from statementDataArray
	statementDataArray = removeColumns(statementDataArray, colIndexShortlist)

	// Duplicate top cells (3 months ended 6 months ended cells) into empty string cells that arose from colspan
	for j := 1; j < len(statementDataArray[0]); j++ {
		if statementDataArray[0][j] == "" {
			statementDataArray[0][j] = statementDataArray[0][j-1]
		}
	}

	// Rearrange array into headers and data format
	statementDataClean := StatementData{
		Headers: make([][]string, len(statementData.Headers)),
		Data:    make([][]string, len(statementDataArray)-len(statementData.Headers)),
	}

	copy(statementDataClean.Headers, statementDataArray[:len(statementData.Headers)])
	copy(statementDataClean.Data, statementDataArray[len(statementData.Headers):])

	//add an empty row at the end of statementDataClean.Headers to separate the headers from the data
	statementDataClean.Headers = append(statementDataClean.Headers, []string{})

	// Create and add accessionNumber, form, and report date rows at the very top of headers
	reportDate, form, err := FindReportDateAndFormGivenAccessionNumber(accessionNumber, client)
	if err != nil {
		log.Printf("Error finding report date and form: %v", err)
		reportDate = ""
		form = ""
	}

	formRow := make([]string, len(statementDataClean.Headers[0]))
	formRow[0] = "form"
	for i := 1; i < len(formRow); i++ {
		formRow[i] = form
	}

	reportDateRow := make([]string, len(statementDataClean.Headers[0]))
	reportDateRow[0] = "reportDate"
	for i := 1; i < len(reportDateRow); i++ {
		reportDateRow[i] = reportDate
	}

	accessionNumberRow := make([]string, len(statementDataClean.Headers[0]))
	accessionNumberRow[0] = "accessionNumber"
	for i := 1; i < len(accessionNumberRow); i++ {
		accessionNumberRow[i] = accessionNumber
	}

	statementDataClean.Headers = append([][]string{accessionNumberRow, formRow, reportDateRow}, statementDataClean.Headers...)

	return statementDataClean
}

// FindReportDateAndFormGivenAccessionNumber finds the report date and form type for a given accession number from MongoDB
func FindReportDateAndFormGivenAccessionNumber(accessionNumber string, client *mongo.Client) (ReportDate string, Form string, err error) {
	collection := utilityfunctions.GetMongoDBCollection(client)

	filter := bson.M{"accessionnumber": accessionNumber}
	var result bson.M
	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return "", "", err
	}
	reportDate, _ := result["reportdate"].(string)
	form, _ := result["form"].(string)

	return reportDate, form, nil
}

// saveParsedRfileAsCSV saves the parsed R file as a CSV
func saveParsedRfileAsCSV(statementData *StatementData, CIK string, accessionNumber string, RfileName string) error {
	re := regexp.MustCompile(`\.(html|htm|xml|xbrl)$`)
	csvFileName := re.ReplaceAllString(RfileName, ".csv")
	outputDirPath := filepath.Join("SEC-files", "filingSummaryAndRfiles", CIK, accessionNumber)
	outputFilePath := filepath.Join(outputDirPath, csvFileName)

	// Ensure that the directory exists
	if err := os.MkdirAll(outputDirPath, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	csvLines := []string{}

	// Add headers to CSV
	for _, headerRow := range statementData.Headers {
		row := []string{}
		for i, cell := range headerRow {
			row = append(row, escapeCsvCell(cell, false, i == 0))
		}
		csvLines = append(csvLines, strings.Join(row, ","))
	}

	// Add data rows to CSV
	for _, dataRow := range statementData.Data {
		row := []string{}
		for i, cell := range dataRow {
			row = append(row, escapeCsvCell(cell, true, i == 0))
		}
		csvLines = append(csvLines, strings.Join(row, ","))
	}

	// Join all lines into a single CSV string
	csvContent := strings.Join(csvLines, "\n")

	// Write CSV string to a file
	err := os.WriteFile(outputFilePath, []byte(csvContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing CSV file: %w", err)
	}

	fmt.Printf("CSV file has been saved to %s\n", outputFilePath)
	return nil
}

func escapeCsvCell(cell string, isDataCell bool, isFirstColumn bool) string {
	if cell == "" {
		return ""
	}

	cellString := cell

	// Clean cell content if it's a data cell but not the first column
	if isDataCell && !isFirstColumn {
		if strings.Contains(cellString, "(") && strings.Contains(cellString, ")") {
			cellString = "-" + cellString // Prepend hyphen if parentheses are present
		}
		// Remove unwanted characters
		cellString = strings.NewReplacer(",", "", "$", "", "(", "", ")", "").Replace(cellString)
	} else {
		// Escape double quotes for CSV
		cellString = strings.ReplaceAll(cellString, "\"", "\"\"")
	}

	// Enclose cell in double quotes if it contains commas, newlines, or quotes
	if strings.ContainsAny(cellString, "\"\n,") {
		cellString = fmt.Sprintf("\"%s\"", cellString)
	}

	return cellString
}

func indexOf(slice []int, value int) int {
	for i, v := range slice {
		if v == value {
			return i // Found the value, return the index
		}
	}
	return -1 // Value not found, return -1
}

func removeColumns(arr [][]string, columnsToDelete []int) [][]string {
	// Create a map for faster lookups
	deleteMap := make(map[int]bool)
	for _, col := range columnsToDelete {
		deleteMap[col] = true
	}

	// Build a new 2D slice without the deleted columns
	newArr := make([][]string, len(arr))
	for i, row := range arr {
		newRow := []string{}
		for j, elem := range row {
			if !deleteMap[j] {
				newRow = append(newRow, elem)
			}
		}
		newArr[i] = newRow
	}

	return newArr
}
