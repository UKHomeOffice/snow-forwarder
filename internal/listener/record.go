package listener

import (
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// Record represents a change event
type Record struct {
	SupplierRef string `json:"supplierRef"`
	Status      string `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Starts      string `json:"startTime"`
	Ends        string `json:"endTime"`
	Table       string
}

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

// PutRec puts an item in DynamoDB
func (d *DB) PutRec(r *Record) (*dynamodb.PutItemOutput, error) {

	av, err := dynamodbattribute.MarshalMap(r)
	if err != nil {
		return nil, err
	}

	if r.SupplierRef == "" {
		return nil, errors.New("missing supplierRef")
	}

	input := &dynamodb.PutItemInput{
		Item:                av,
		TableName:           aws.String(r.Table),
		ConditionExpression: aws.String("attribute_not_exists(supplierRef)"),
	}

	out, err := d.DynamoDB.PutItem(input)
	if err != nil {
		return nil, err
	}

	log.Printf("added %v - %v to table %v", r.SupplierRef, r.Status, r.Table)
	return out, nil
}

// UpdateRec updates an item in DynamoDB
func (d *DB) UpdateRec(r *Record) (*dynamodb.UpdateItemOutput, error) {

	if r.SupplierRef == "" {
		return nil, errors.New("missing supplierRef")
	}

	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(r.Table),
		UpdateExpression: aws.String("SET #S = :cst"),
		ExpressionAttributeNames: map[string]*string{
			"#S": aws.String("status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":cst": {
				S: aws.String(r.Status),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"supplierRef": {
				S: aws.String(r.SupplierRef),
			},
		},
	}
	out, err := d.DynamoDB.UpdateItem(input)
	if err != nil {
		return nil, err
	}

	log.Printf("updated %v with %v on table %v", r.SupplierRef, r.Status, r.Table)
	return out, nil
}

// recorder handles DynamoDB ops
func recorder(r *Record) error {

	db, err := newDB()
	if err != nil {
		return err
	}
	_, err = db.PutRec(r)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				log.Println("item exists, will try to update instead")
				_, err = db.UpdateRec(r)
				if err != nil {
					return err
				}
			case dynamodb.ErrCodeInternalServerError:
				return err
			default:
				return aerr
			}
		} else {
			return err
		}
	}
	return nil
}
