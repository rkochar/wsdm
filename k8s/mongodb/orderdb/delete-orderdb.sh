#!/bin/bash

echo "Deleting orderdb sts"
kubectl delete -f ./k8s/mongodb/orderdb/

echo "Deleting orderdb pvc"
kubectl delete pvc -l 'app in (orderdb-0, orderdb-1, orderdb-2, orderdb-3, orderdb-4)'
