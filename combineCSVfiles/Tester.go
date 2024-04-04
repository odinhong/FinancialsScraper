package combinecsvfiles

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func Tester(CIK string, client *mongo.Client) {
	BSfilepaths, ISfilepaths, CISfilepaths, CFfilepaths, err := GetCSVfilepathsInOrder(CIK, client)

	//need to work with csv files from here
	fmt.Println("BSfilepaths:", BSfilepaths, "ISfilepaths:", ISfilepaths, "CISfilepaths:", CISfilepaths, "CFfilepaths:", CFfilepaths, err)
}
