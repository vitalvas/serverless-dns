package main

import "github.com/aws/aws-lambda-go/lambda"

type Server struct {
	dnsHandler *DNSHandler
}

func main() {
	srv := Server{}

	lambda.Start(srv.lambdaHandler)
}
