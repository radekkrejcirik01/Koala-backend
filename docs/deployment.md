## Deployment of Golang app to AWS's Lambda

1. export AWS_PROFILE=radek
2. export AWS_REGION=eu-central-1
3. GOARCH=amd64 GOOS=linux go build main.go
4. zip -r koala.zip . -x '*.git*'

## Create
aws lambda create-function --function-name koala --zip-file fileb://koala.zip --handler main --runtime go1.x --role "arn:aws:iam::832685173872:role/lambda-basic-execution"

## Update
aws lambda update-function-code --function-name koala --zip-file fileb://koala.zip

## Invoke
aws lambda invoke --function-name koala --invocation-type "RequestResponse" response.txt

## Migrations
Migrations don't run on every cold start, they run only when `MIGRATE=true` is set.
After deploying a change of models, enable it, invoke the function once and disable it again:

aws lambda update-function-configuration --function-name koala --environment "Variables={...,MIGRATE=true}"
