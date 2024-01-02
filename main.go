package main

import (
	"log"

	fetchdata "github.com/Programmerdin/FinancialDataSite_Go/fetchData"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var downloadLink string = "https://www.sec.gov/Archives/edgar/daily-index/bulkdata/submissions.zip"
	var filePath string = "SECfiles/submissions/submissions.zip"
	fetchdata.DownloadOneSECFile(downloadLink, filePath)
}
