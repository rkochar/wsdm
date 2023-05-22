#!/bin/bash

./mongo/deploy-mongo-stateful-replicasets.sh

./kafka/deploy-kafka.sh

./k8s/deploy-microservice.sh
