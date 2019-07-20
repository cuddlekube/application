.PHONY: build run deploy feed-flow-up feed-flow-down

build:
	@bin/build.sh

run:
	@docker-compose up

deploy:
	@bin/deploy.sh

feed-flow-up:
	@docker-compose up -d dynamolocal
	@docker-compose up -d register-api
	@docker-compose up -d list-api
	@docker-compose up -d happiness-api
	@docker-compose up -d feed-api
	@docker-compose ps
	# this step makes it non-idempotent
	@bin/local-dynamo-init.sh

feed-flow-down:
	@docker-compose down
