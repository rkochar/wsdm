#!/bin/bash


echo "Local builder"
eval $(minikube -p wsdm2 docker-env)


docker build . --build-arg SERVICE=lockmaster -t lockmaster:latest
echo "Deleting lockmaster"
kubectl delete -f k8s/microservices/lockmaster-app.yaml

echo "Deleting lockmaster"
kubectl apply -f k8s/microservices/lockmaster-app.yaml