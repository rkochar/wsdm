#1/bin/bash
echo $(pwd)
echo "Deploying stock"
kubectl apply -f ./k8s/stock-app.yaml

echo "Deploying order"
kubectl apply -f ./k8s/order-app.yaml

echo "Deploying payment"
kubectl apply -f ./k8s/payment-app.yaml


