#!/bin/bash

./k8s/delete-microservice.sh

./mongo/delete-mongo-stateful-replicasets.sh

./kafka/delete-kafka.sh
