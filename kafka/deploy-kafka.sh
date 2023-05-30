#1/bin/bash

echo "Creating namespace"
kubectl apply -f ./kafka/namespace.yaml

echo "Deploying kafka and zookeeper"
kubectl apply -f ./kafka/zookeeper.yaml
sleep 5
kubectl apply -f ./kafka/kafka.yaml
