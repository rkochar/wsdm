#!/bin/bash

echo "Deploying orderdb"
kubectl apply -f ./k8s/microservices/api-gateway/
