#!/bin/bash

echo "Deploying orderdb"
kubectl apply -f ./mongo/orderdb/
