#!/bin/bash

echo "Deleting paymentdb sts"
kubectl delete -f ./k8s/mongodb/paymentdb/

echo "Deleting paymentdb pvc"
kubectl delete pvc -l 'app in (paymentdb-0, paymentdb-1, paymentdb-2)'

