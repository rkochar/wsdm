#!/bin/bash

./mongo/delete-mongo-stateful-replicasets.sh

./k8s/delete-microservice.sh

./kafka/delete-kafka.sh
