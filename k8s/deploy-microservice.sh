#1/bin/bash
echo $(pwd)
echo "Deploying stock"
kubectl apply -f ./k8s/stock-app.yaml

echo "Deploying order"
kubectl apply -f ./k8s/order-app.yaml

echo "Deploying payment"
kubectl apply -f ./k8s/payment-app.yaml

echo "Deploying ingress"
kubectl apply -f ./k8s/ingress-service.yaml

echo "Deploying lockmaster"
kubectl apply -f ./k8s/sql-pv.yaml
kubectl apply -f ./k8s/lockmaster-sql-deployment.yaml
