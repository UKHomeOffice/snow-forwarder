package listener

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

func (md *mockDynamoDB) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	output := new(dynamodb.PutItemOutput)
	return output, md.err
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

func TestPutRec(t *testing.T) {

	tt := []struct {
		name        string
		supplierRef string
		status      string
		title       string
		description string
		starts      string
		ends        string
		table       string
		err         string
	}{
		{name: "good", supplierRef: "abc-123", status: "in progress", title: "upgrade", description: "lorem ipsum", starts: "01-01-1970", ends: "02-01-1970", table: "foo"},
		{name: "bad", err: "missing supplierRef"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			putter := new(DB)
			putter.DynamoDB = &mockDynamoDB{}

			if tc.err == "" {
				rec := Record{
					SupplierRef: tc.supplierRef,
					Status:      tc.status,
					Title:       tc.title,
					Description: tc.description,
					Starts:      tc.starts,
					Ends:        tc.ends,
					Table:       tc.table,
				}
				_, err := putter.PutRec(&rec)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			rec := Record{}
			_, err := putter.PutRec(&rec)
			if msg := err.Error(); !strings.Contains(msg, tc.err) {
				t.Errorf("expected error %q, got: %q", tc.err, msg)
			}
		})
	}
}

func TestUpdateRec(t *testing.T) {

	tt := []struct {
		name        string
		supplierRef string
		status      string
		err         string
	}{
		{name: "good", supplierRef: "abc-123", status: "in progress"},
		{name: "bad", err: "missing supplierRef"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			updater := new(DB)
			updater.DynamoDB = &mockDynamoDB{}

			if tc.err == "" {
				rec := Record{
					SupplierRef: tc.supplierRef,
					Status:      tc.status,
				}
				_, err := updater.UpdateRec(&rec)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
			rec := Record{}
			_, err := updater.UpdateRec(&rec)
			if msg := err.Error(); !strings.Contains(msg, tc.err) {
				t.Errorf("expected error %q, got: %q", tc.err, msg)
			}
		})
	}
}
