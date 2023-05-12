# SetUp minikube

Start a cluster called wsdm.

`minikube start -p wsdm`

Enable ingress (will likely need it later).

`minikube addons enable ingress`

Install redis.

`./../deploy-charts-minikube.sh`

Authenticate with ghcr (do this once, set your pat to expire in 60 days). Kube [docs](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry) if needed.

`kubectl create secret docker-registry regcred-ghcr --docker-server=https://ghcr.io --docker-username=<username> --docker-password=<pat> --docker-email=<email>`

Deploy stock, order and payment.

`kubectl apply -f <microservice>-app.yaml`

Deploy ingress

`kubectl apply -f ingress-service.yaml`

Delete a deployment.

`kubectl delete -f <name>.yaml`

Delete redis.

`helm list --all`
`helm delete <name>`

If you want to edit a deployment or other kube files, edit the file in this directory and apply it. You can also edit the file inside Kubernetes but that is not a good idea because the changes will be on Kubernetes but the deliverables are the yaml files (drifts are bad).
