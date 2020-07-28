package notifier

import (
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// DB wraps DynamodDB with iface pkg for easier testing
type DB struct {
	DynamoDB dynamodbiface.DynamoDBAPI
}

func newDB() (*DB, error) {

	var db = new(DB)
	reg, ok := os.LookupEnv("REGION")
	if !ok {
		return nil, errors.New("missing AWS region")
	}

	awsConfig := aws.Config{
		Region: aws.String(reg),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            awsConfig,
	}))

	svc := dynamodb.New(sess, aws.NewConfig())
	db.DynamoDB = dynamodbiface.DynamoDBAPI(svc)
	return db, nil
}

// AddID adds internal_identifier to existing db record
func (d *DB) AddID(ur *Response) (*dynamodb.UpdateItemOutput, error) {

	tab, ok := os.LookupEnv("TABLE_NAME")
	if !ok {
		return nil, errors.New("missing table name")
	}

	if ur.SupplierRef == "" {
		return nil, errors.New("missing supplierRef")
	}

	// create update payload
	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(tab),
		UpdateExpression: aws.String("SET internal_identifier = :cid"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":cid": {
				S: aws.String(ur.IntIdent),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"supplierRef": {
				S: aws.String(ur.SupplierRef),
			},
		},
	}

	out, err := d.DynamoDB.UpdateItem(input)
	if err != nil {
		return nil, err
	}

	log.Printf("added change ref: %v, to %v on table %v", ur.IntIdent, ur.SupplierRef, tab)

	return out, nil
}
