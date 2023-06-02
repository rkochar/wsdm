#!/bin/bash

echo "Deploying paymentdb"
kubectl apply -f ./mongo/paymentdb/
