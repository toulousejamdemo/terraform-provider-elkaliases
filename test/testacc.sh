#!/bin/bash

docker compose -f test/docker-compose.yml up -d --wait
TF_ACC=1 ELKALIASES_URL=http://localhost:9200 ELKALIASES_TOKEN=empty go test ./provider -v
ret=$(echo $?)
docker compose -f test/docker-compose.yml down

exit $ret
