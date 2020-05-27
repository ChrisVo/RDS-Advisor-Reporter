package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/support"
)

var reportName string

func init() {
	flag.StringVar(&reportName, "filename", "report.csv", "Name of report")
	flag.StringVar(&reportName, "f", "report.csv", "Name of report")
}

// GetAccountID returns the AWS Account ID we're operating in
func GetAccountID(sess *session.Session) string {
	stsSvc := sts.New(sess)
	accountID, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	if err != nil {
		log.Fatal("Couldn't get the caller identity")
	}

	return *accountID.Account
}

// WriteCsv will write the line item parameter to the csv file.
func WriteCsv(lineItem []*support.TrustedAdvisorResourceDetail, accountID string) {

	flag.Parse()

	if _, err := os.Stat(reportName); os.IsNotExist(err) {
		fmt.Println(reportName + " doesn't exist. Creating the report.")
		_, err := os.Create(reportName)
		if err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.OpenFile(reportName,
		os.O_APPEND|os.O_WRONLY, os.ModeAppend|0777)
	if err != nil {
		log.Fatal(err)
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(file)

	defer file.Close()
	for _, item := range lineItem {
		var sliceItem []string
		for _, data := range item.Metadata {
			sliceItem = append(sliceItem, *data)
		}
		// Add account ID to the line item
		sliceItem = append(sliceItem, accountID)
		csvWriter.Write(sliceItem)
	}
	csvWriter.Flush()
	fmt.Printf("RDS Utilization report Trusted Advisor report created.\n")
}

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	accountID := GetAccountID(sess)
	if err != nil {
		log.Fatal(err)
	}

	// Trusted Advisor service
	supportSvc := support.New(sess)
	checks, err := supportSvc.DescribeTrustedAdvisorChecks(
		&support.DescribeTrustedAdvisorChecksInput{
			Language: aws.String("en"),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	var checkID string
	for _, check := range checks.Checks {
		if *check.Category == "cost_optimizing" &&
			*check.Name == "Amazon RDS Idle DB Instances" {
			checkID = *check.Id
		}
	}

	input := &support.DescribeTrustedAdvisorCheckResultInput{
		CheckId: aws.String(checkID),
	}

	summary, err := supportSvc.DescribeTrustedAdvisorCheckResult(input)
	if err != nil {
		log.Fatal(err)
	}

	WriteCsv(summary.Result.FlaggedResources, accountID)

}
