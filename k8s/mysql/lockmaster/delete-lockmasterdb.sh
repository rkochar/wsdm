#!/bin/bash

echo "Deploying lockmasterdb"
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-0.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-1.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-2.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-3.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-4.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-5.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-6.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-7.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-8.yaml
kubectl delete -f ./k8s/mysql/lockmaster/lockmasterdb-9.yaml

echo "Deleting lockmasterdb pvc"
kubectl delete pvc -l 'app in (lockmasterdb-0, lockmasterdb-1, lockmasterdb-2, lockmasterdb-3, lockmasterdb-4, lockmasterdb-5, lockmasterdb-6, lockmasterdb-7, lockmasterdb-8, lockmasterdb-9)'
