#!/bin/bash

set -eu

img=kinesis-streams-agent-logrotate
dir=$(pwd)

echo -n 'go build ... '
cd ../../ && GOOS=linux GOARCH=amd64 go build && cd "$dir"
cp ../../kinesis-streams-agent ./
echo 'finished.'

docker build . -t "$img"
docker run -d --rm --privileged "$img" /sbin/init
container_id=$(docker ps | grep "$img" | awk '{print $1'})
docker exec -it "$container_id" /bin/bash

docker rm -f "$container_id"
