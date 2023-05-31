#!/bin/bash

echo "Deleting mongo"
./k8s/mongodb/delete-mongodb.sh

./k8s/microservices/delete-microservice.sh

./k8s/kafka/delete-kafka.sh
