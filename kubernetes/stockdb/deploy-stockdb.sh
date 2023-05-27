#!/bin/bash

echo "Deploying stockdb"
kubectl apply -f ./kubernetes/stockdb/
