#!/bin/bash


docker-compose up --build

./mongodb/deploy-mongo.sh

./kafka/deploy-kakfa.sh

./k8s/deploy-microservice.sh
