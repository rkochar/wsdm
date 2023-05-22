#!/bin/bash

echo "Deleting mongodb-stock"
kubectl delete -f ./mongo/stock/
kubectl delete pvc -l app=mongodb-stock

echo "Deleting mongodb-order"
kubectl delete -f ./mongo/order/
kubectl delete pvc -l app=mongodb-order

echo "Deleting mongodb-payment"
kubectl delete -f ./mongo/payment/
kubectl delete pvc -l app=mongodb-payment

# Make pv available again
echo "Releasing pv"
kubectl get pv | tail -n+2 | awk '$5 == "Released" {print $1}' | xargs -I{} kubectl patch pv {} --type='merge' -p '{"spec":{"claimRef": null}}'
