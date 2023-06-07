# Web-scale Data Management Project Template

Basic project structure with Python's Flask and Redis. 
**You are free to use any web framework in any language and any database you like for this project.**

### Project structure

* `k8s`
    Folder containing the kubernetes deployments, apps and services for the ingress, order, payment and stock services.
    
* `src`
    Folder containing application and business logic.

* `test`
    Folder containing some basic correctness tests for the entire system.

* `config`
    Folder storing config files.

## Actual Kubernetes

```
./deploy-cluster.sh
kubectl port-forward service/nginx-service 8080:80
```

Use a seperate terminal and keep it open for port-forwarding


To cleanup,

`./delete-cluster.sh`
