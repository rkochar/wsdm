#!/bin/bash


echo "Reset"
eval $(minikube -p wsdm1 docker-env)
docker build ./payment/ -t user:latest

echo "Deleting payment"
kubectl delete -f ./k8s/payment-app.yaml

echo "Deleting payment"
kubectl apply -f ./k8s/payment-app.yaml
