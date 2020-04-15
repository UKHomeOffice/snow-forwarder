package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// handler waits for a DynamaDB stream event and calls notifier
func handler(e events.DynamoDBEvent) error {

	for _, record := range e.Records {
		fmt.Printf("processing request data for event ID %s, type %s.\n", record.EventID, record.EventName)

		msg := struct {
			Issue       string `json:"Issue"`
			Status      string `json:"Status"`
			Summary     string `json:"Summary"`
			Description string `json:"Description"`
			Component   string `json:"Component"`
			Starts      string `json:"Starts"`
			Ends        string `json:"Ends"`
		}{}

		msg.Issue = record.Change.NewImage["Issue"].String()
		msg.Starts = record.Change.NewImage["Starts"].String()
		msg.Ends = record.Change.NewImage["Ends"].String()
		msg.Status = record.Change.NewImage["Status"].String()
		msg.Summary = record.Change.NewImage["Summary"].String()
		msg.Component = record.Change.NewImage["Component"].String()
		msg.Description = record.Change.NewImage["Description"].String()

		mBytes, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("could not construct payload: %s", err)
		}

		err = notifier(mBytes)
		if err != nil {
			return fmt.Errorf("could not make HTTP request: %s", err)
		}
	}
	return nil
}

// notifier calls SNOW API
func notifier(m []byte) error {

	p := bytes.NewReader(m)

	user := os.Getenv("SNOW_USERNAME")
	pass := os.Getenv("SNOW_PASSWORD")

	u, err := url.Parse(os.Getenv("SNOW_URL"))
	if err != nil {
		return fmt.Errorf("could not parse SNOW endpoint URL: %s", err)
	}

	req, err := http.NewRequest("POST", u.String(), p)
	if err != nil {
		return fmt.Errorf("could not form request: %s", err)
	}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %s", err)
	}

	defer resp.Body.Close()

	fmt.Printf("sent request, SNOW replied: %s", resp.Body)

	return nil
}

func main() {

	lambda.Start(handler)

}
