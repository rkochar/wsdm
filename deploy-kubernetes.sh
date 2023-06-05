#!/usr/bin/env bash

docker build . --build-arg SERVICE=api-gateway  -t api-gateway:latest

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx

helm repo update

helm install -f helm-config/nginx-helm-values.yaml nginx ingress-nginx/ingress-nginx

./deploy-minikube.sh

echo "Starting port-forward"
kubectl port-forward service/nginx-ingress-nginx-controller 8080:80
