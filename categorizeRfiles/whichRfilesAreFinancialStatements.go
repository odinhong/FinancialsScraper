package categorizefinancialstatements

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/xmlpath.v2"
)

type RfileFinancialStatementObject struct {
	FinancialStatementType string
	FileName               string
	LongName               string
	ShortName              string
	MenuCategory           string
}

func CategorizeRfilesOfFinancialStatementsFromFilingSummaryXML(filePath string) ([]RfileFinancialStatementObject, error) {
	XMLdata, err := ReadFilingSummaryXmlFile(filePath)
	if err != nil {
		fmt.Printf("error CategorizeRfilesOfFinancialStatementsFromFilingSummaryXML function: %v", err)
		return nil, err
	}

	root, err := xmlpath.Parse(strings.NewReader(string(XMLdata)))
	if err != nil {
		fmt.Printf("error parsing XML: %v\n", err)
		return nil, err
	}

	HtmlFileNameTagPath := xmlpath.MustCompile("/FilingSummary/MyReports/Report[1]/HtmlFileName")
	XmlFileNameTagPath := xmlpath.MustCompile("/FilingSummary/MyReports/Report[1]/XmlFileName")
	ShortNameTagPath := xmlpath.MustCompile("/FilingSummary/MyReports/Report[1]/ShortName")
	doesHtmlFileNameTagExist := HtmlFileNameTagPath.Exists(root)
	doesXmlFileNameTagExist := XmlFileNameTagPath.Exists(root)
	doesShortNameTagExist := ShortNameTagPath.Exists(root)

	doBothHtmlAndXmlExistOrDontExist := (doesHtmlFileNameTagExist && doesXmlFileNameTagExist) || (!doesHtmlFileNameTagExist && !doesXmlFileNameTagExist)
	whichFileNameTagToUse := ""
	if doBothHtmlAndXmlExistOrDontExist {
		return nil, fmt.Errorf("this FilingSummaryFile has both HtmlFileName and XmlFileName or Don't have either")
	} else {
		if doesHtmlFileNameTagExist {
			whichFileNameTagToUse = "HtmlFileName"
		} else {
			whichFileNameTagToUse = "XmlFileName"
		}
	}

	BS_struct := RfileFinancialStatementObject{}
	IS_struct := RfileFinancialStatementObject{}
	CIS_struct := RfileFinancialStatementObject{}
	CF_struct := RfileFinancialStatementObject{}

	reportsPath := xmlpath.MustCompile("/FilingSummary/MyReports/Report")
	fileNamePath := xmlpath.MustCompile(whichFileNameTagToUse)
	longNamePath := xmlpath.MustCompile("LongName")
	shortNamePath := xmlpath.MustCompile("ShortName")
	menuCategoryPath := xmlpath.MustCompile("MenuCategory")

	reportIter := reportsPath.Iter(root)
	for reportIter.Next() {
		reportNode := reportIter.Node()

		if longName, ok := longNamePath.String(reportNode); ok {
			financialStatementTypeFrom1stfilter := whichFinancialStatement1stFilter(longName)
			if financialStatementTypeFrom1stfilter != "" {
				confirmationfrom2ndfilter := whichFinancialStatement2ndFilter(longName, financialStatementTypeFrom1stfilter)
				if confirmationfrom2ndfilter {
					switch financialStatementTypeFrom1stfilter {
					case "BS":
						BS_struct.FinancialStatementType = "BS"
						BS_struct.FileName, _ = fileNamePath.String(reportNode)
						BS_struct.LongName = longName
						BS_struct.MenuCategory, _ = menuCategoryPath.String(reportNode)
						if doesShortNameTagExist {
							BS_struct.ShortName, _ = shortNamePath.String(reportNode)
						}
					case "IS":
						IS_struct.FinancialStatementType = "IS"
						IS_struct.FileName, _ = fileNamePath.String(reportNode)
						IS_struct.LongName = longName
						IS_struct.MenuCategory, _ = menuCategoryPath.String(reportNode)
						if doesShortNameTagExist {
							IS_struct.ShortName, _ = shortNamePath.String(reportNode)
						}
					case "CIS":
						CIS_struct.FinancialStatementType = "CIS"
						CIS_struct.FileName, _ = fileNamePath.String(reportNode)
						CIS_struct.LongName = longName
						CIS_struct.MenuCategory, _ = menuCategoryPath.String(reportNode)
						if doesShortNameTagExist {
							CIS_struct.ShortName, _ = shortNamePath.String(reportNode)
						}
					case "CF":
						CF_struct.FinancialStatementType = "CF"
						CF_struct.FileName, _ = fileNamePath.String(reportNode)
						CF_struct.LongName = longName
						CF_struct.MenuCategory, _ = menuCategoryPath.String(reportNode)
						if doesShortNameTagExist {
							CF_struct.ShortName, _ = shortNamePath.String(reportNode)
						}
					}

					// if all 4 structs are filled then stop the loop
					if BS_struct.FileName != "" && IS_struct.FileName != "" && CIS_struct.FileName != "" && CF_struct.FileName != "" {
						fmt.Println("break")
						break
					}
				}
			}
		}
	}
	return []RfileFinancialStatementObject{BS_struct, IS_struct, CIS_struct, CF_struct}, nil
}

func ReadFilingSummaryXmlFile(filePath string) (string, error) {
	XMLfile, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		return "", err
	}
	defer XMLfile.Close()

	XMLdata, err := io.ReadAll(XMLfile)
	if err != nil {
		fmt.Printf("error reading file: %v", err)
		return "", err
	}

	return string(XMLdata), nil
}
