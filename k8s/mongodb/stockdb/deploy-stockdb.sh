#!/bin/bash

echo "Deploying stockdb"
kubectl apply -f ./k8s/mongodb/stockdb/
