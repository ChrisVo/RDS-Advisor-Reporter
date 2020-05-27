package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/support"
)

var reportName string

func init() {
	flag.StringVar(&reportName, "filename", "report.csv", "Name of report")
	flag.StringVar(&reportName, "filename", "report.csv", "Name of report")
}

// WriteCsv will write the line item parameter to the csv file.
func WriteCsv(lineItem []*support.TrustedAdvisorResourceDetail) {

	flag.Parse()
	fmt.Println(reportName)
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
	csvWriter.Write([]string{"region", "db_name", "multi_az",
		"instance_type",
		"storage_provision",
		"days_since_last_connection",
		"estimated_monthly_savings"})
	for _, item := range lineItem {
		var sliceItem []string
		for _, data := range item.Metadata {
			sliceItem = append(sliceItem, *data)
		}
		csvWriter.Write(sliceItem)
	}
	csvWriter.Flush()
	fmt.Println("RDS Utilization report Trusted Advisor report created. See " + reportName)
}

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})

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

	WriteCsv(summary.Result.FlaggedResources)

}
