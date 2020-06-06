package notifier

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestSetMsg(t *testing.T) {

	tt := []struct {
		name   string
		event  string
		status string
		expect string
	}{
		{name: "create", event: "INSERT", status: "Scheduled", expect: "HO_SIAM_IN_REST_CHG_POST_JSON"},
		{name: "update", event: "MODIFY", status: "In Progress", expect: "HO_SIAM_IN_REST_CHG_UPDATE_JSON"},
		{name: "delete", event: "REMOVE", expect: ""},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			av := make(map[string]events.DynamoDBAttributeValue)
			val := events.NewStringAttribute(tc.status)
			av["status"] = val

			change := &events.DynamoDBStreamRecord{
				NewImage: av,
			}

			event := &events.DynamoDBEventRecord{
				EventName: tc.event,
				Change:    *change,
			}

			p := Payload{}
			msg, err := p.SetMsg(event)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if msg.MessageID != tc.expect {
				t.Errorf("expected %v, got %v", tc.expect, msg.MessageID)

			}
		})
	}
}
