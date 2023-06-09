# Web-scale Data Management Project Template

Uses Go, Kafka, MySQL and MongoDB.

5 static mongodb databases are made for order, stock and payment. Order, stock and payment autoscale from 5 to 25. id are hashed and distributed equally among the 5 databases (that is why they are static).

Lockmaster (should be called orchestrator/ses) has 10 MySQL databases and also autoscales from 5 to 25.

### Project structure

* `k8s`
    Folder containing the kubernetes deployments, apps and services for the ingress, order, payment and stock services.
    
* `src`
    Folder containing application and business logic.

* `test`
    Folder containing some basic correctness tests for the entire system.

## Actual Kubernetes

Zookeeper can take some time to start up (if other pods start before zookeeper, they will complain).
Wait for cluster to normalise and setup.

```
./deploy-cluster.sh
kubectl port-forward service/nginx-service 8080:80
```

Use a seperate terminal and keep it open for port-forwarding


To cleanup,

`./delete-cluster.sh`

__This is for course IN4331: Web-scale Data Management__
Slides are located at ./slides.pdf

