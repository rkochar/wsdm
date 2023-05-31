#!/bin/bash

echo "Delete kafka"
docker-compose down --build
kubectl delete -f zookeeper-deployment.yaml
kubectl delete -f kafka-deployment.yaml