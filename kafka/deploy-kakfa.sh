#!/bin/bash

echo "Deploying kafka"
cd kafka
docker-compose up --build
kubectl apply -f zookeeper-deployment.yaml
kubectl apply -f kafka-deployment.yaml