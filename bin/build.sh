#!/bin/bash

set -e

commit_sha=$(git rev-list -1 HEAD)
app_version=0.0.1

for dir in feed-api list-api order-api register-api validate-api public-site happiness-api
do
	if [ -d "${dir}" ]
	then
		cd ${dir}
		docker build . -t ${dir} --build-arg=app_version=${app_version} --build-arg=commit_sha=${commit_sha}
		docker tag ${dir} ${dir}:${commit_sha}
		cd ../
    fi
done
