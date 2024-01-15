package parserfiles

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xmlquery"
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
		ParseOneRfileAndSaveAsCSV(CIK, accessionNumbers_to_parse[i], Rfilenames_to_parse[i])
	}
}

func ParseOneRfileAndSaveAsCSV(CIK, accessionNumber, RfileName string) error {
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
		fmt.Println(parsedRfile)
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

	cleanParsedRfile := CleanParsedRfile(&parsedRfile)

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
	filePath := filepath.Join("SEC-files/companyFilingFiles", CIK, accessionNumber, RfileName)
	xmlBytes, err := os.ReadFile(filePath)
	if err != nil {
		return statementData, err
	}

	doc, err := xmlquery.Parse(strings.NewReader(string(xmlBytes)))
	if err != nil {
		return statementData, err
	}

	// Iterate over each <Column> element
	labelArrays := [][]string{}
	columns := xmlquery.Find(doc, "//Column")
	for _, column := range columns {
		labels := xmlquery.Find(column, "Label")
		for i, label := range labels {
			text := label.SelectAttr("Label")
			for len(labelArrays) <= i {
				labelArrays = append(labelArrays, []string{})
			}
			labelArrays[i] = append(labelArrays[i], text)
		}
	}

	statementData.Headers = append(statementData.Headers, labelArrays...)

	// Find <ReportName> or <ReportLongName> element and add it to headers[0][0]
	reportName := xmlquery.FindOne(doc, "//ReportName")
	reportLongName := xmlquery.FindOne(doc, "//ReportLongName")
	if reportName != nil {
		statementData.Headers[0] = append([]string{reportName.InnerText()}, statementData.Headers[0]...)
	} else if reportLongName != nil {
		statementData.Headers[0] = append([]string{reportLongName.InnerText()}, statementData.Headers[0]...)
	}

	// Add empty cells to beginning of each headers array
	for i := 1; i < len(statementData.Headers); i++ {
		statementData.Headers[i] = append([]string{""}, statementData.Headers[i]...)
	}

	// Iterate over each <Row> element
	rows := xmlquery.Find(doc, "//Row")
	for _, row := range rows {
		rowData := []string{}
		elementName := xmlquery.FindOne(row, "ElementName").InnerText()
		rowData = append(rowData, elementName)
		cells := xmlquery.Find(row, "Cell")
		for _, cell := range cells {
			numericAmount := xmlquery.FindOne(cell, "NumericAmount").InnerText()
			rowData = append(rowData, numericAmount)
		}
		statementData.Data = append(statementData.Data, rowData)
	}

	// Remove rows that are abstract rows in statementData.Data
	for i := 0; i < len(statementData.Data); i++ {
		isLineItemAbstract := strings.Contains(statementData.Data[i][0], "Abstract")
		isLineItem0OrEmpty := statementData.Data[i][1] == "0" || statementData.Data[i][1] == ""
		if isLineItemAbstract && isLineItem0OrEmpty {
			statementData.Data = append(statementData.Data[:i], statementData.Data[i+1:]...)
			i--
		}
	}

	return statementData, nil
}

func CleanParsedRfile(statementData *StatementData) StatementData {
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

	// Duplicate some cells into empty string cells that arose from colspan
	for i := 0; i < 3 && i < len(statementDataArray); i++ {
		for j := 1; j < len(statementDataArray[i]); j++ {
			if statementDataArray[i][j] == "" {
				statementDataArray[i][j] = statementDataArray[i][j-1]
			}
		}
	}

	// Rearrange array into headers and data format
	statementDataClean := StatementData{
		Headers: make([][]string, len(statementData.Headers)),
		Data:    make([][]string, len(statementDataArray)-len(statementData.Headers)),
	}

	copy(statementDataClean.Headers, statementDataArray[:len(statementData.Headers)])
	copy(statementDataClean.Data, statementDataArray[len(statementData.Headers):])

	return statementDataClean
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
