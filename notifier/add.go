package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// AddIntID adds internal_identifier to DynamoDB item
func (ur *res) AddIntID(svc *dynamodb.DynamoDB) error {

	tab := os.Getenv("TABLE_NAME")

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

	// update item on DynamoDB
	_, err := svc.UpdateItem(input)
	if err != nil {
		return err
	}

	log.Printf("added change ref " + ur.IntIdent + " to " + ur.SupplierRef + " - " + " on table " + tab)

	return nil
}
