#!/bin/bash

echo "Deploying paymentdb"
kubectl apply -f ./k8s/mongodb/paymentdb/
