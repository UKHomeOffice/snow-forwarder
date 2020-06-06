package notifier

import (
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockDynamoDB struct {
	dynamodbiface.DynamoDBAPI
	err error
}

func (md *mockDynamoDB) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	output := new(dynamodb.UpdateItemOutput)
	return output, md.err
}

func TestNewDB(t *testing.T) {

	tt := []struct {
		name string
		err  string
	}{
		{name: "good"},
		{name: "bad", err: "missing AWS region"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			if tc.err != "" {
				os.Unsetenv("REGION")
				_, err := newDB()
				if msg := err.Error(); !strings.Contains(msg, tc.err) {
					t.Errorf("expected error %q, got: %q", tc.err, msg)
				}
			}
			os.Setenv("REGION", "eu")
			_, err := newDB()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddID(t *testing.T) {

	tt := []struct {
		name        string
		supplierRef string
		internalID  string
		err0        string
		err1        string
	}{
		{name: "good", supplierRef: "abc-123", internalID: "ch-123"},
		{name: "bad", supplierRef: "abc-123", err0: "missing table name", err1: "missing supplierRef"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			updater := new(DB)
			updater.DynamoDB = &mockDynamoDB{}

			if tc.err0 != "" {
				os.Setenv("TABLE_NAME", "bar")
				rec := Response{
					SupplierRef: tc.supplierRef,
					IntIdent:    tc.internalID,
				}
				_, err := updater.AddID(&rec)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			os.Unsetenv("TABLE_NAME")
			rec := Response{}
			_, err := updater.AddID(&rec)
			if msg := err.Error(); !strings.Contains(msg, tc.err0) {
				t.Errorf("expected error %q, got: %q", tc.err0, msg)
			}

			os.Setenv("TABLE_NAME", "bar")
			_, err = updater.AddID(&rec)
			if msg := err.Error(); !strings.Contains(msg, tc.err1) {
				t.Errorf("expected error %q, got: %q", tc.err1, msg)
			}
		})
	}
}
