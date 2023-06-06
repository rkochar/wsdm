#!/bin/bash


echo "Local builder"
eval $(minikube -p wsdm2 docker-env)


docker build . --build-arg SERVICE=payment -t ghcr.io/rkochar/wsdm/payment:latest
docker build . --build-arg SERVICE=order -t ghcr.io/rkochar/wsdm/order:latest
docker build . --build-arg SERVICE=stock -t ghcr.io/rkochar/wsdm/stock:latest
docker build . --build-arg SERVICE=lockmaster -t ghcr.io/rkochar/wsdm/lockmaster:latest
docker build . --build-arg SERVICE=api-gateway -t ghcr.io/rkochar/wsdm/api-gateway:latest
docker build ./src/nginx -t ghcr.io/rkochar/wsdm/nginx:latest