#!/bin/bash

echo "Deploying mongodb-stock"
kubectl apply -f ./mongo/stock/

echo "Deploying mongodb-order"
kubectl apply -f ./mongo/order/

echo "Deploying mongodb-payment"
kubectl apply -f ./mongo/payment/
