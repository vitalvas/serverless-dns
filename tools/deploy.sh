#!/bin/bash

set -e -x

STACK_NAME='serverless-dns'

AWS_REGION='eu-central-1'
AWS="aws --profile vitalvas --region ${AWS_REGION}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-w -s' -o main *.go
zip function.zip main

${AWS} cloudformation package \
    --template-file tools/template.yml \
    --s3-bucket vv-eu-central-1-lambda-code \
    --s3-prefix ${STACK_NAME} \
    --region ${AWS_REGION} \
    --output-template-file package.yml

${AWS} cloudformation deploy \
    --template-file package.yml \
    --region ${AWS_REGION} \
    --capabilities CAPABILITY_IAM \
    --stack-name ${STACK_NAME}

rm -f main function.zip
