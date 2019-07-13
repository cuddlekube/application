.PHONY: cfn-lint deploy-logs-global deploy-vpc-demo

build:
	@bin/build.sh

run:
	@docker-compose up

deploy:
	@bin/deploy.sh