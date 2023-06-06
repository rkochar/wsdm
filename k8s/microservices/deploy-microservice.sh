#1/bin/bash

echo "Deploying ingress"
kubectl apply -f ./k8s/microservices/ingress-service.yaml

echo "Deploying stock"
kubectl apply -f ./k8s/microservices/stock-app.yaml

echo "Deploying order"
kubectl apply -f ./k8s/microservices/order-app.yaml

echo "Deploying payment"
kubectl apply -f ./k8s/microservices/payment-app.yaml

sleep 10
echo "Deploying lockmaster"
kubectl apply -f ./k8s/microservices/lockmaster-app.yaml

echo "Deploying API Gateway"
./k8s/microservices/api-gateway/deploy-api-gateway.sh

echo "Deleting NGINX"
kubectl apply -f ./k8s/microservices/nginx/nginx.yaml
