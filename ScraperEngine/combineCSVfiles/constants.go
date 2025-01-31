package combinecsvfiles

// Balance sheet category names
var BalanceSheetCategoryNames = []string{
	"Current Assets",
	"Total Current Assets",
	"Non Current Assets",
	"Total Assets",
	"Current Liabilities",
	"Total Current Liabilities",
	"Non Current Liabilities",
	"Total Liabilities",
	"Stockholders' Equity",
	"Total Stockholders' Equity",
	"Other Equities",
}

// Row indices for various financial report metadata
var (
    AccessionNumberRowIndex = 0
    FormRowIndex           = 1
    ReportDateRowIndex     = 2
    DenominationRowIndex   = 4
    ReportPeriodRowIndex   = 5
    ReportDurationRowIndex = 6
    SeparatorRowIndex      = 7
)

// You can add more constant slices or maps here as needed
