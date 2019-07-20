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

Dynamo Schema

```
- ckid -- HASH
- name -- String
- type -- String (aws server classes?)
- service -- int (e.g 20 years in service)
- happiness -- int (1 being shit 10 being super happy)
- petname -- String 
- os -- String (linux, windows)
- image -- String 
```
Curl request to register

```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"ckid":"ck1","name":"foo foo cuddly poops", "type": "m2.xlarge","service":10,"happiness":1, "petname":"sven", "os":"linux", "image":"https://media.giphy.com/media/a5ptfHj2GqOmk/giphy.gif"}' \
  http://localhost:8083/
```


## Feed API Flow

To standup containers for feed flow with dynamo initialised run the following

```
make feed-flow-up
```
p.s the above is not idempotent yet

to tear it down run:

```
make feed-flow-down
```




