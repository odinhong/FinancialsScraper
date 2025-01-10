package categorizefinancialstatements

import (
	"strings"
)

var common_BS_names = []string{"Balance Sheet", "Financial Position"}
var common_IS_names = []string{"Income Statement", "Statements of Income", "Statement of Income", "Statements of Operation", "Statement of Operation", "Statements of Operations and Comprehensive", "Statements of Operation and Comprehensive", "Statement of Operations and Comprehensive", "Statement of Operation and Comprehensive"}
var common_CIS_names = []string{"Statements of Comprehensive Income", "Statement of Comprehensive Income", "Comprehensive Income", "COMPREHENSIVE LOSS"}
var common_CF_names = []string{"Statements of Cash Flows", "Statement of Cash Flows", "Statement of Cash Flow"}

var common_BS_exclusion_terms = []string{"Parenthetical", "Derivative", "Fair", "Current", "Detail", "Disclosure"}
var common_IS_exclusion_terms = []string{"Detail", "Notes"}
var common_CIS_exclusion_terms = []string{"Detail", "Disclosure", "Notes"}
var common_CF_exclusion_terms = []string{"Detail", "Notes"}

var clean_common_BS_names = []string{}
var clean_common_IS_names = []string{}
var clean_common_CIS_names = []string{}
var clean_common_CF_names = []string{}

var clean_common_BS_exclusion_terms = []string{}
var clean_common_IS_exclusion_terms = []string{}
var clean_common_CIS_exclusion_terms = []string{}
var clean_common_CF_exclusion_terms = []string{}

func init() {
	clean_common_BS_names = cleanNames(common_BS_names)
	clean_common_IS_names = cleanNames(common_IS_names)
	clean_common_CIS_names = cleanNames(common_CIS_names)
	clean_common_CF_names = cleanNames(common_CF_names)

	clean_common_BS_exclusion_terms = cleanNames(common_BS_exclusion_terms)
	clean_common_IS_exclusion_terms = cleanNames(common_IS_exclusion_terms)
	clean_common_CIS_exclusion_terms = cleanNames(common_CIS_exclusion_terms)
	clean_common_CF_exclusion_terms = cleanNames(common_CF_exclusion_terms)
}

func cleanNames(names []string) []string {
	cleaned := make([]string, 0, len(names)) // Preallocate with same capacity for efficiency
	for _, name := range names {
		cleaned = append(cleaned, cleanString(name))
	}
	return cleaned
}

func whichFinancialStatement1stFilter(s string) string {
	// Clean s
	s_clean := cleanString(s)

	// Return "BS" "IS" "CIS" "CF" given s
	for _, name := range clean_common_BS_names {
		if strings.Contains(s_clean, name) {
			return "BS"
		}
	}
	for _, name := range clean_common_IS_names {
		if strings.Contains(s_clean, name) {
			return "IS"
		}
	}
	for _, name := range clean_common_CIS_names {
		if strings.Contains(s_clean, name) {
			return "CIS"
		}
	}
	for _, name := range clean_common_CF_names {
		if strings.Contains(s_clean, name) {
			return "CF"
		}
	}
	return "" // Return "" if no match is found
}

func whichFinancialStatement2ndFilter(longName string, financialStatementTypeFrom1stfilter string) bool {
	longName_clean := cleanString(longName)
	switch {
	case financialStatementTypeFrom1stfilter == "BS":
		for _, term := range clean_common_BS_exclusion_terms {
			if strings.Contains(longName_clean, term) {
				return false
			}
		}
	case financialStatementTypeFrom1stfilter == "IS":
		for _, term := range clean_common_IS_exclusion_terms {
			if strings.Contains(longName_clean, term) {
				return false
			}
		}
	case financialStatementTypeFrom1stfilter == "CIS":
		for _, term := range clean_common_CIS_exclusion_terms {
			if strings.Contains(longName_clean, term) {
				return false
			}
		}
	case financialStatementTypeFrom1stfilter == "CF":
		for _, term := range clean_common_CF_exclusion_terms {
			if strings.Contains(longName_clean, term) {
				return false
			}
		}
	}
	return true
}

func cleanString(s string) string {
	// Lowercase the string
	s = strings.ToLower(s)

	// Remove spaces from the string
	s = strings.ReplaceAll(s, " ", "")

	return s
}
