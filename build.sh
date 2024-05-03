#!/bin/bash

REGISTRY='192.168.103.220/library'
go build -o ./build/
docker build -t $REGISTRY/knative-agent:latest .
docker push $REGISTRY/knative-agent:latest
