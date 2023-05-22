#1/bin/bash

echo "Deleting kafka and zookeeper"
kubectl delete -f ./kafka/zookeeper.yaml
kubectl delete -f ./kafka/kafka.yaml

echo "Deleting namespace"
kubectl delete -f ./kafka/namespace.yaml
