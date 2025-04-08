#!/bin/bash

docker compose -f test/docker-compose.yml up -d --wait
curl -X PUT "localhost:9200/index?pretty"
TF_ACC=1 ELASTICSEARCH_ENDPOINT=http://localhost:9200 ELASTICSEARCH_API_KEY=empty go test ./provider -v
ret=$(echo $?)
docker compose -f test/docker-compose.yml down

exit $ret
