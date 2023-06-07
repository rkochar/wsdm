#!/bin/bash

echo "Deleting stockdb sts"
kubectl delete -f ./k8s/mongodb/stockdb/

echo "Deleting stockdb pvc"
kubectl delete pvc -l 'app in (stockdb-0, stockdb-1, stockdb-2, stockdb-3, stockdb-4)'
