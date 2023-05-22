#1/bin/bash

echo "Deleting stock"
kubectl delete -f ./k8s/stock-app.yaml

echo "Deleting order"
kubectl delete -f ./k8s/order-app.yaml

echo "Deleting payment"
kubectl delete -f ./k8s/payment-app.yaml

echo "Deleting ingress"
kubectl delete -f ./k8s/ingress-service.yaml

echo "Deleting lockmaster"
kubectl delete -f ./k8s/lockmaster-sql-deployment.yaml

echo "Release SQL pv"
kubectl delete pvc mysql-pvc
kubectl get pv | tail -n+2 | awk '$5 == "Released" {print $1}' | xargs -I{} kubectl patch pv {} --type='merge' -p '{"spec":{"claimRef": null}}'
