#1/bin/bash

echo "Deleting kafka and zookeeper"
kubectl delete -f ./k8s/kafka/zookeeper.yaml
kubectl delete -f ./k8s/kafka/kafka.yaml

echo "Deleting namespace"
kubectl delete -f ./k8s/kafka/namespace.yaml
