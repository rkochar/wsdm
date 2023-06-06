#!/bin/bash

echo "Starting kafka deployment"
./k8s/kafka/deploy-kafka.sh

echo "Starting mongo deployment"
./k8s/mongodb/deploy-mongodb.sh

echo "Starting mysql deployment"
./k8s/mysql/lockmaster/deploy-lockmasterdb.sh

sleep 10
echo "Starting microservices deployment"
./k8s/microservices/deploy-microservice.sh
