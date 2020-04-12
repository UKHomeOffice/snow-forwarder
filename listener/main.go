package main

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/apex/gateway"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/tidwall/gjson"
)

// Message is the struct for the SNOW payload
type Message struct {
	Issue       string
	Status      string
	Summary     string
	Description string
	Component   string
	Starts      string
	Ends        string
}

// chkEmpty checks for null strings
func chkEmpty(ss ...string) bool {
	for _, s := range ss {
		if s == "" {
			return true
		}
	}
	return false
}

// handler waits for a payload from JSD and calls writer
func handler(rw http.ResponseWriter, req *http.Request) {

	// read incoming payload request body
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(rw, "could not read request body", http.StatusBadRequest)
	}
	newStr := buf.String()

	// get relevant fields as envars
	// TODO handle multiple components
	com := gjson.Get(newStr, os.Getenv("COMPONENT_FIELD"))
	des := gjson.Get(newStr, os.Getenv("DESCRIPTION_FIELD"))
	fti := gjson.Get(newStr, os.Getenv("FINISH_TIME_FIELD"))
	iid := gjson.Get(newStr, os.Getenv("ISSUE_ID_FIELD"))
	sta := gjson.Get(newStr, os.Getenv("STATUS_FIELD"))
	sti := gjson.Get(newStr, os.Getenv("START_TIME_FIELD"))
	sum := gjson.Get(newStr, os.Getenv("SUMMARY_FIELD"))

	if chkEmpty(iid.Str, sum.Str, com.Str, des.Str, sta.Str, sti.Str, fti.Str) {
		http.Error(rw, "missing value in incoming payload", http.StatusBadRequest)
		log.Fatal("missing value in incoming payload")
	} else {
		log.Printf("change webhook received: %s ", iid)
	}

	msg := Message{
		Issue:       iid.Str,
		Status:      sta.Str,
		Summary:     sum.Str,
		Description: des.Str,
		Component:   com.Str,
		Starts:      sti.Str,
		Ends:        fti.Str,
	}

	if !writer(msg) {
		rw.WriteHeader(http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusOK)
}

// writer puts the message in DynamoDB
func writer(msg Message) bool {

	tab := os.Getenv("TABLE_NAME")
	reg := os.Getenv("REGION")

	// create a session and service with Dynamodb
	awsConfig := aws.Config{
		Region: aws.String(reg),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            awsConfig,
	}))

	svc := dynamodb.New(sess, aws.NewConfig())

	// marshall into attribute value map
	av, err := dynamodbattribute.MarshalMap(msg)
	if err != nil {
		log.Fatal("could not marshall message: ", err)
		return false
	}

	// create message in specified table
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tab),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		log.Fatal("could not write message to db: ", err)
		return false
	}

	log.Printf("Successfully added '" + msg.Issue + "-" + msg.Status + " to table " + tab)
	return true
}

func main() {

	http.HandleFunc("/", handler)
	log.Fatal(gateway.ListenAndServe("", nil))

}
