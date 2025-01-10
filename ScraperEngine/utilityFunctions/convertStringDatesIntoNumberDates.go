package utilityfunctions

import (
	"fmt"

	"github.com/araddon/dateparse"
)

// convert string dates like  Oct. 1, 2021 to number dates like 20211001
func ConvertDateStringToYYYYMMDD(dateString string) string {
	ParsedDate, err := dateparse.ParseStrict(dateString)
	if err != nil {
		fmt.Println("Could not parse date:", dateString)
	}

	//convert date to first yyyyMMdd in string format
	StringParsedDate := fmt.Sprintf("%d%02d%02d", ParsedDate.Year(), int(ParsedDate.Month()), ParsedDate.Day())

	return StringParsedDate
}
