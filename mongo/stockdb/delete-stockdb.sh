#!/bin/bash

echo "Deleting stockdb sts"
kubectl delete -f ./kubernetes/stockdb/

echo "Deleting stockdb pvc"
kubectl delete pvc -l 'app in (stockdb-0, stockdb-1, stockdb-2)'

