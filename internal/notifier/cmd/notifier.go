package main

import (
	"github.com/UKHomeOffice/snow-forwarder/internal/notifier"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {

	lambda.Start(notifier.Handler)
}
