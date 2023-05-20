#1/bin/bash

echo "Deleting kafka and zookeeper"
kubectl delete -f ./kafka/
