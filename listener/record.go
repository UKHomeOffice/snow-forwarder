package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var tab string = os.Getenv("TABLE_NAME")

// PutRec puts an item in DynamoDB
func (r *rec) PutRec(svc *dynamodb.DynamoDB) error {

	av, err := dynamodbattribute.MarshalMap(r)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:                av,
		TableName:           aws.String(tab),
		ConditionExpression: aws.String("attribute_not_exists(supplierRef)"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return err
	}

	log.Printf("added " + r.SupplierRef + " - " + r.Status + " to table " + tab)
	return nil
}

// UpdateRec updates an item in DynamoDB
func (r *rec) UpdateRec(svc *dynamodb.DynamoDB) error {

	input := &dynamodb.UpdateItemInput{
		TableName:        aws.String(tab),
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
	_, err := svc.UpdateItem(input)
	if err != nil {
		return err
	}

	log.Printf("updated " + r.SupplierRef + " with " + r.Status + " on table " + tab)
	return nil
}
