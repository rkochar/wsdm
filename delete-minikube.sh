#!/bin/bash

echo "Deleting mongo"
./mongo/delete-mongo-stateful-replicasets.sh

./k8s/delete-microservice.sh

./kafka/delete-kafka.sh
