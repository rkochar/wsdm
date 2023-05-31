#!/bin/bash

echo "Deleting orderdb sts"
kubectl delete -f ./kubernetes/orderdb/

echo "Deleting orderdb pvc"
kubectl delete pvc -l 'app in (orderdb-0, orderdb-1, orderdb-2)'

