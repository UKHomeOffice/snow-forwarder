package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/tidwall/gjson"
)

type rec struct {
	SupplierRef string `json:"supplierRef"`
	Status      string `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
}

// Handler receives a payload from JSD and writes to DynamoDB
func Handler(rw http.ResponseWriter, req *http.Request) {

	// read incoming request body
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		log.Fatalf("could not read JSD request body: %v", err)
	}
	newStr := buf.String()

	des := gjson.Get(newStr, os.Getenv("DESCRIPTION_FIELD"))
	fti := gjson.Get(newStr, os.Getenv("FINISH_TIME_FIELD"))
	iid := gjson.Get(newStr, os.Getenv("ISSUE_ID_FIELD"))
	sta := gjson.Get(newStr, os.Getenv("STATUS_FIELD"))
	sti := gjson.Get(newStr, os.Getenv("START_TIME_FIELD"))
	sum := gjson.Get(newStr, os.Getenv("SUMMARY_FIELD"))

	// check no value is missing
	val := []string{iid.Str, sta.Str, sum.Str, des.Str, sti.Str, fti.Str}
	for _, s := range val {
		if s == "null" {
			log.Fatal("missing value in JSD request")
		}
	}
	log.Printf("processing JSD event: %v, status: %v\n", iid, sta)

	// format timestamps
	const (
		layout     = "2006-01-02T15:04:05.000+0000"
		layoutSNOW = "2006-01-02 15:04:05"
	)

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatalf("could not set SNOW timezone: %v", err)
	}

	ftt, err := time.Parse(layout, fti.Str)
	if err != nil {
		log.Fatalf("could not parse finish time: %v", err)
	}

	stt, err := time.Parse(layout, sti.Str)
	if err != nil {
		log.Fatalf("could not parse start time: %v", err)
	}

	// add JSD ticket link to description
	desc := "\nFor the most up-to-date info, visit " +
		os.Getenv("JSD_URL") + "/" + iid.Str + "\n" + des.Str

	r := rec{
		SupplierRef: iid.Str,
		Status:      sta.Str,
		Title:       sum.Str,
		Description: desc,
		StartTime:   stt.In(loc).Format(layoutSNOW),
		EndTime:     ftt.In(loc).Format(layoutSNOW),
	}

	// configure aws session
	reg := os.Getenv("REGION")

	awsConfig := aws.Config{
		Region: aws.String(reg),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            awsConfig,
	}))

	svc := dynamodb.New(sess, aws.NewConfig())

	err = r.PutRec(svc)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Println("item exists, will try to update instead")
				err = r.UpdateRec(svc)
				if err != nil {
					log.Fatalf("could not update db record: %v", err)
				}
			case dynamodb.ErrCodeInternalServerError:
				log.Fatalf("could not create db record: %v", err)
			default:
				log.Fatalf(aerr.Error())
			}
		} else {
			log.Fatalf(err.Error())
		}
	}

	rw.WriteHeader(http.StatusOK)
}
