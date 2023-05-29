#!/bin/bash

echo "Deploying orderdb"
kubectl apply -f ./kubernetes/orderdb/
