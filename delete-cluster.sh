#!/bin/bash

echo "Start deleting mongodb"
./k8s/mongodb/delete-mongodb.sh

echo "Start deleting mongodb"
./k8s/mysql/lockmaster/delete-lockmasterdb.sh

echo "Start deleting microservices"
./k8s/microservices/delete-microservice.sh

echo "Start deleting kafka"
./k8s/kafka/delete-kafka.sh

