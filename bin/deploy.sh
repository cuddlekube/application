#!/bin/bash

set -e
$(aws ecr get-login --no-include-email --region ap-southeast-2)

commit_sha=$(git rev-list -1 HEAD)
app_version=1.0.0

for dir in dummy-passthrough-api feed-api list-api order-api register-api validate-api public-site
do
	if [ -d "$dir" ]
	then
		docker tag ${dir}:latest 183741349056.dkr.ecr.ap-southeast-2.amazonaws.com/${dir}:latest
        docker push 183741349056.dkr.ecr.ap-southeast-2.amazonaws.com/${dir}:latest
    fi
done