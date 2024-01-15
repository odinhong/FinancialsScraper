package geteverythinggivencik

import (
	categorizefinancialstatements "github.com/Programmerdin/FinancialDataSite_Go/categorizeRfiles"
	fetchdata "github.com/Programmerdin/FinancialDataSite_Go/fetchDataFolder"
	parserfiles "github.com/Programmerdin/FinancialDataSite_Go/parseRfiles"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetEverythingGivenCIK(CIK string, client *mongo.Client) {
	fetchdata.Store10K10QmetadataFromSubmissionFilesCIKtoMongoDB(CIK, client)
	fetchdata.CheckAllFilingIndexJsonForExistenceOfFilingSummary(CIK, client)
	fetchdata.DownloadFilingSummaryFiles(CIK, client)

	categorizefinancialstatements.ParseManyFilingSummaryXmlFilesAndSaveToMongoGivenCIK(CIK, client)
	parserfiles.DownloadRfiles(CIK, client)
	parserfiles.ParseManyRfilesAndSaveAsCSVs(CIK, client)
}
