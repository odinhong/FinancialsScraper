package main

import (
	"fmt"
	"log"

	//import fetchdata package here
	"github.com/Programmerdin/FinancialDataSite_Go/fetchdata"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// var KO_CIK string = "0000021344"
	// var META_CIK string = "0001326801"
	// var AAPL_CIK string = "0000320193"
	var SMRT_CIK string = "0001837014"

	// test, err := fetchdata.GetSubmissionFilesOfCIK(SMRT_CIK)
	// // fmt.Println(test)

	// test2, err := fetchdata.ReadJsonFile(test[0])
	// // fmt.Println(test2)

	// test3 := gjson.Get(test2, "filings.recent.accessionNumber")
	// // fmt.Println(test3)
	// fmt.Println(reflect.TypeOf(test3))

	// test4 := gjson.Parse(test2)
	// // fmt.Println(test4)
	// fmt.Println(reflect.TypeOf(test4))

	var CIK_submission_filePath string = "SEC-files/submissions/CIK" + SMRT_CIK + ".json"
	test5, err := fetchdata.ParseOneJsonSubmissionFile(CIK_submission_filePath)
	fmt.Println(test5, err)
}
