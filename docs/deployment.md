## Deployment of Golang app to AWS's Lambda

The function runs on the `provided.al2023` runtime with the `arm64` architecture,
the binary must be named `bootstrap`.

```sh
export AWS_PROFILE=radek
export AWS_REGION=eu-central-1
```

## Build
```sh
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap main.go
zip koala-arm.zip bootstrap fcm-client.json
```

## Update
```sh
aws lambda update-function-code --function-name koala --zip-file fileb://koala-arm.zip
aws lambda wait function-updated --function-name koala
```

## Create
```sh
aws lambda create-function --function-name koala --zip-file fileb://koala-arm.zip --handler bootstrap --runtime provided.al2023 --architectures arm64 --timeout 60 --role "arn:aws:iam::832685173872:role/lambda-basic-execution"
aws lambda put-function-concurrency --function-name koala --reserved-concurrent-executions 40
```

Reserved concurrency limits running instances, each instance holds one DB connection.
Keep it about 15 below the DB's `max_connections` (`SHOW VARIABLES LIKE 'max_connections';`).

## Invoke
```sh
aws lambda invoke --function-name koala --invocation-type "RequestResponse" response.txt
```

## Logs
```sh
aws logs tail /aws/lambda/koala --since 10m
```

## Migrations
Migrations don't run on every cold start, they run only when `MIGRATE=true` is set.
Needed only after a change of models (a new table, column or index), do it at a quiet time.

1. Allow only one instance, so instances don't run the migration in parallel
   ```sh
   aws lambda put-function-concurrency --function-name koala --reserved-concurrent-executions 1
   ```
2. Set `MIGRATE` to `true` in the console: Lambda → koala → Configuration → Environment variables → Edit.
   Don't use `update-function-configuration --environment`, it replaces all environment variables.
3. Invoke the function and check the logs for errors
   ```sh
   aws lambda invoke --function-name koala response.txt
   aws logs tail /aws/lambda/koala --since 5m
   ```
4. Set `MIGRATE` back to `false` and restore the concurrency
   ```sh
   aws lambda put-function-concurrency --function-name koala --reserved-concurrent-executions 40
   ```
