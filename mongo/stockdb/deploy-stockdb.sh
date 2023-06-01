#!/bin/bash

echo "Deploying stockdb"
kubectl apply -f ./mongo/stockdb/
