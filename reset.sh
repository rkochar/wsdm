#!/bin/bash


echo "Local builder"
eval $(minikube -p wsdm2 docker-env)


docker build . --build-arg SERVICE=order -t order:latest
echo "Deleting order"
kubectl delete -f k8s/microservices/order-app.yaml

echo "Deleting order"
kubectl apply -f k8s/microservices/order-app.yaml