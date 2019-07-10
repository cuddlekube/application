#!/bin/bash 

set -e

commit_sha=$(git rev-list -1 HEAD)
app_version=1.0.0

for dir in dummy-passthrough-api feed-api list-api order-api register-api validate-api
do 
	if [ -d "$dir" ]
	then
		cd $dir
		docker build . -t $dir --build-arg=app_version=$app_version --build-arg=commit_sha=$commit_sha
		cd ../ 
    fi
done