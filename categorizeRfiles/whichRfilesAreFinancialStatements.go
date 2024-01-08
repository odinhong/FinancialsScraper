package categorizefinancialstatements

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/xmlpath.v2"
)

func ReadFilingSummaryFile(filePath string) error {
	XMLfile, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		return err
	}
	defer XMLfile.Close()

	XMLdata, err := io.ReadAll(XMLfile)
	if err != nil {
		fmt.Printf("error reading file: %v", err)
		return err
	}

	path := xmlpath.MustCompile("/FilingSummary/MyReports/Report[2]/LongName")
	root, err := xmlpath.Parse(strings.NewReader(string(XMLdata)))
	if err != nil {
		fmt.Printf("error parsing XML: %v\n", err)
		return err
	}

	if value, ok := path.String(root); ok {
		fmt.Printf("The LongName of the second Report is: %s\n", value)
	} else {
		fmt.Println("Couldn't find the LongName in the second Report.")
	}

	return nil
}
