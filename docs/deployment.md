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

## WebSocket API
The same `koala` function serves the WebSocket API (typing, presence and live chat updates),
it sends events back through `WS_ENDPOINT`. The app's side is described in `docs/websocket.md`.

1. Create the API, the Lambda integration, the routes and the stage
   ```sh
   API_ID=$(aws apigatewayv2 create-api --name koala-ws --protocol-type WEBSOCKET \
     --route-selection-expression '$request.body.action' --query ApiId --output text)

   INTEGRATION_ID=$(aws apigatewayv2 create-integration --api-id $API_ID --integration-type AWS_PROXY \
     --integration-uri arn:aws:apigateway:eu-central-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-central-1:832685173872:function:koala/invocations \
     --query IntegrationId --output text)

   for ROUTE in '$connect' '$disconnect' '$default' 'typing'; do
     aws apigatewayv2 create-route --api-id $API_ID --route-key "$ROUTE" --target integrations/$INTEGRATION_ID
   done

   aws apigatewayv2 create-stage --api-id $API_ID --stage-name prod --auto-deploy
   ```
2. Allow the API to invoke the function
   ```sh
   aws lambda add-permission --function-name koala --statement-id koala-ws \
     --action lambda:InvokeFunction --principal apigateway.amazonaws.com \
     --source-arn "arn:aws:execute-api:eu-central-1:832685173872:$API_ID/*"
   ```
3. Allow the function to send to the connections
   ```sh
   aws iam put-role-policy --role-name lambda-basic-execution --policy-name koala-ws-connections \
     --policy-document "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Action\":\"execute-api:ManageConnections\",\"Resource\":\"arn:aws:execute-api:eu-central-1:832685173872:$API_ID/prod/POST/@connections/*\"}]}"
   ```
4. Set `WS_ENDPOINT` to `https://$API_ID.execute-api.eu-central-1.amazonaws.com/prod` in the console,
   same as `MIGRATE` in Migrations. Without it the function works, it only sends no socket events.
5. Run the migration, it creates the `connections` table
6. The app connects to `wss://$API_ID.execute-api.eu-central-1.amazonaws.com/prod`

WebSocket invocations share the reserved concurrency and DB connections with the REST API.
