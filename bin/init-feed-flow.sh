#!/bin/bash

set -e

# Set dummy creds used in apps
export AWS_ACCESS_KEY_ID=123
export AWS_SECRET_ACCESS_KEY=123
export AWS_DEFAULT_REGION=ap-southeast-2

# create dynamodb table 
aws dynamodb create-table \
    --table-name cuddlykube \
    --attribute-definitions AttributeName=ckid,AttributeType=S \
    --key-schema AttributeName=ckid,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=10,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000 \
    --region ap-southeast-2


# create one item via the register api
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"ckid":"ck1","name":"foo foo cuddly poops", "type": "m2.xlarge","service":10,"happiness":1, "petname":"sven", "os":"linux", "image":"https://media.giphy.com/media/a5ptfHj2GqOmk/giphy.gif"}' \
  http://localhost:8083/