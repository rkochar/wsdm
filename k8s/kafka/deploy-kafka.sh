#1/bin/bash

#echo "Creating namespace"
#kubectl apply -f ./k8s/kafka/namespace.yaml

echo "Deploying zookeeper"
kubectl apply -f ./k8s/kafka/zookeeper.yaml
sleep 5

echo "Deploying kafka"
kubectl apply -f ./k8s/kafka/kafka.yaml
