# SetUp minikube

Start a cluster called wsdm.

`minikube start -p wsdm`

Enable ingress (will likely need it later).

`minikube addons enable ingress`

Deploy microservices and ingresses.

`./deploy-microservice.sh`

Delete microservices and ingresses.

`./delete-microservice.sh`

If you want to edit a deployment or other kube files, edit the file in this directory and apply it `kubectl replace -f <file.yaml>`. You can also edit the file inside Kubernetes but that is not a good idea because the changes will be on Kubernetes but the deliverables are the yaml files (drifts are bad).
