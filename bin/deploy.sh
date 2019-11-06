#!/bin/bash
set -e
AWS_ACCOUNT="YOURACCOUNTID"

$(aws ecr get-login --no-include-email --region ap-southeast-2)

commit_sha=$(git rev-list -1 HEAD)
app_version=0.0.3

for dir in feed-api list-api order-api register-api validate-api public-site happiness-api
do
	if [ -d "$dir" ]
	then
		docker tag ${dir}:latest ${AWS_ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${dir}:latest
		docker tag ${dir}:latest ${AWS_ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${dir}:${commit_sha}
        docker push ${AWS_ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${dir}:latest
        docker push ${AWS_ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${dir}:${commit_sha}
    fi
done
