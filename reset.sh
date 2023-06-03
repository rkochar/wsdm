#!/bin/bash


echo "Local builder"
eval $(minikube -p wsdm1 docker-env)


docker build . --build-arg SERVICE=stock -t stock:latest
echo "Deleting stock"
kubectl delete -f k8s/microservices/stock-app.yaml

echo "Deleting stock"
kubectl apply -f k8s/microservices/stock-app.yaml
