# SetUp minikube

Start a cluster called wsdm.

`minikube start -p wsdm`

Enable ingress (will likely need it later).

`minikube addons enable ingress`

Install redis.

`./../deploy-charts-minikube.sh`

Deploy stock, order and user.

`kubectl apply -f <microservice>-app.yaml`

Deploy ingress

`kubectl apply -f ingress-service.yaml`

Delete a deployment.

`kubectl delete -f <name>.yaml`
