#!/bin/bash

echo "Deploying lockmasterdb"
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-0.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-1.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-2.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-3.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-4.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-5.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-6.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-7.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-8.yaml
kubectl apply -f ./k8s/mysql/lockmaster/lockmasterdb-9.yaml
