#!/bin/bash


echo "Local builder"
eval $(minikube -p wsdm1 docker-env)


docker build . --build-arg SERVICE=payment -t payment:latest
echo "Deleting payment"
kubectl delete -f k8s/microservices/payment-app.yaml

echo "Deleting payment"
kubectl apply -f k8s/microservices/payment-app.yaml
