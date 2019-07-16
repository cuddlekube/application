# ccau-2019-apps
Applications to be used for the CCAU demo


Boiler plate golang based apis with `/` and `/version` endpoints. 


Build all docker images:

```
bin/build.sh
```

Run all the images:

```
docker-compose up
```


Local dynamo db

```
docker-compose up dynamolocal
```

Create a dynamo table

```
aws dynamodb create-table \
    --table-name cuddlykube \
    --attribute-definitions AttributeName=ckid,AttributeType=S \
    --key-schema AttributeName=ckid,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=10,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000 \
    --region ap-southeast-2
```

Curl request to register

```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"ckid":"ck1","name":"foo foo cuddly poops"}' \
  http://localhost:8080/
```