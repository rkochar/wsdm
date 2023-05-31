#!/bin/bash


echo "Reset"
eval $(minikube -p wsdm1 docker-env)
#docker build ./order/ -t payment:latest

echo "Deleting order"
kubectl delete -f ./k8s/payment-app.yaml

echo "Deleting order"
kubectl apply -f ./k8s/payment-app.yaml
