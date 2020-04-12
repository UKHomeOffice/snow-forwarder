package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type pld struct {
	SupplierRef string `json:"supplierRef"`
	Status      string `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Success     bool   `json:"success,omitempty"`
}

type msg struct {
	MessageID string `json:"messageid"`
	IntID     string `json:"internal_identifier,omitempty"`
	Payload   pld    `json:"payload"`
}

type res struct {
	SupplierRef string `json:"supplierRef"`
	IntIdent    string `json:"internal_identifier"`
}

func (p *pld) SetMsg(record *events.DynamoDBEventRecord) (*msg, error) {

	var m msg

	// ignore remove events
	if record.EventName == "REMOVE" {
		log.Printf("ignoring DynamoDB event ID %s, as it is of type %s.\n", record.EventID, record.EventName)
		return &m, nil
	}
	log.Printf("processing DynamoDB event ID %s, type %s.\n", record.EventID, record.EventName)

	// construct payloads
	if record.Change.NewImage["status"].String() == "In Progress" || record.Change.NewImage["status"].String() == "Completed" {
		p.Success = true
		m = msg{
			MessageID: "HO_SIAM_IN_REST_CHG_UPDATE_JSON",
			IntID:     record.Change.NewImage["internal_identifier"].String(),
			Payload:   *p,
		}
		return &m, nil
	} else if record.EventName == "INSERT" && record.Change.NewImage["status"].String() == "Scheduled" {
		m = msg{
			MessageID: "HO_SIAM_IN_REST_CHG_POST_JSON",
			Payload:   *p,
		}
		return &m, nil
	} else {
		log.Printf("ignoring event for %s, status: %s\n", record.Change.NewImage["supplierRef"].String(), record.Change.NewImage["status"].String())
		return &m, nil
	}
}

// Handler receives a DynamaDB stream and calls Notify and UpdateRec
func Handler(ctx context.Context, e events.DynamoDBEvent) error {

	for _, record := range e.Records {

		// get relevant values from stream event
		p := pld{
			SupplierRef: record.Change.NewImage["supplierRef"].String(),
			Status:      record.Change.NewImage["status"].String(),
			Title:       record.Change.NewImage["title"].String(),
			Description: record.Change.NewImage["description"].String(),
			StartTime:   record.Change.NewImage["startTime"].String(),
			EndTime:     record.Change.NewImage["endTime"].String(),
		}

		m, err := p.SetMsg(&record)
		if err != nil {
			return err
		}

		if m.MessageID == "" {
			log.Println("event ignored, exiting")
			return nil
		}

		// call Notify and expect internal_identifer in return
		intid, err := m.Notify()
		if err != nil {
			log.Printf("could not call notify: %v", err)
			return err
		}

		if intid == "" {
			log.Printf("notify didn't return a new Change ID, exiting")
			return nil
		}

		// configure session
		reg := os.Getenv("REGION")

		awsConfig := aws.Config{
			Region: aws.String(reg),
		}

		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config:            awsConfig,
		}))

		svc := dynamodb.New(sess, aws.NewConfig())

		// add internal_identifier to db record
		ur := res{
			SupplierRef: p.SupplierRef,
			IntIdent:    intid,
		}

		err = ur.AddIntID(svc)
		if err != nil {
			log.Printf("could not update db with internal identifier: %v", err)
			return err
		}
	}
	return nil
}
