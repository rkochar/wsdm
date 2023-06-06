#1/bin/bash

echo "Deleting stock"
kubectl delete -f ./k8s/microservices/stock-app.yaml

echo "Deleting order"
kubectl delete -f ./k8s/microservices/order-app.yaml

echo "Deleting payment"
kubectl delete -f ./k8s/microservices/payment-app.yaml

echo "Deleting ingress"
kubectl delete -f ./k8s/microservices/ingress-service.yaml

echo "Deleting lockmaster"
kubectl delete -f ./k8s/microservices/lockmaster-app.yaml

echo "Deleting API Gateway"
./k8s/microservices/api-gateway/delete-api-gateway.sh

echo "Deleting NGINX"
kubectl delete -f ./k8s/microservices/nginx/nginx.yaml
