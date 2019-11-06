# ccau-2019-apps

Applications to used for the CCAU Cuddle Kube demo

## Build and run locally

Build all docker images:

```bash
bin/build.sh
# or
make build
```

Run all the images locally:

```bash
docker-compose up
# or
make run
```

### Local dynamo db

```
docker-compose up dynamolocal
```

Create a dynamo table

```bash
aws dynamodb create-table \
    --table-name cuddlykube \
    --attribute-definitions AttributeName=ckid,AttributeType=S \
    --key-schema AttributeName=ckid,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=10,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000 \
    --region ap-southeast-2
```

Dynamo Schema

```ini
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

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"ckid":"ck1","name":"foo foo cuddly poops", "type": "m2.xlarge","service":10,"happiness":1, "petname":"sven", "os":"linux", "image":"https://media.giphy.com/media/a5ptfHj2GqOmk/giphy.gif"}' \
  http://localhost:8083/
```

### Feed API Flow

To standup containers for feed flow with dynamo initialised run the following

```bash
make feed-flow-up
```

p.s the above is not idempotent yet

to tear it down run:

```bash
make feed-flow-down
```

## Deploy

The deploy script will deploy to ECR and assumes you have the repositories created in the ap-southeast-2 region. You will also need to adjust the **AWS_ACCOUNT** variable in the script.

```bash
bin/deploy.sh
# or
make deploy
```
