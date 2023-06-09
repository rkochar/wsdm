#1/bin/bash


echo "Deploying API Gateway"
./k8s/microservices/api-gateway/deploy-api-gateway.sh

echo "Deploying stock"
kubectl apply -f ./k8s/microservices/stock-app.yaml

echo "Deploying order"
kubectl apply -f ./k8s/microservices/order-app.yaml

echo "Deploying payment"
kubectl apply -f ./k8s/microservices/payment-app.yaml

sleep 10
echo "Deploying lockmaster"
kubectl apply -f ./k8s/microservices/lockmaster-app.yaml

sleep 30
echo "Deploying nginx"
kubectl apply -f ./k8s/microservices/nginx.yaml

