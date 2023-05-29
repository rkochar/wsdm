#!/bin/bash

echo "Deploying paymentdb"
kubectl apply -f ./kubernetes/paymentdb/
