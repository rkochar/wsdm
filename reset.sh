#!/bin/bash


echo "Reset"
eval $(minikube -p wsdm1 docker-env)

echo "Deleting order"
kubectl delete -f k8s/microservices/order-app.yaml

echo "Deleting order"
kubectl apply -f k8s/microservices/order-app.yaml
