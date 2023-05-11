#!/usr/bin/env bash

helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

#helm install payment-sql oci://registry-1.docker.io/bitnamicharts/postgresql
#helm install stock-sql oci://registry-1.docker.io/bitnamicharts/postgresql
#helm install order-sql oci://registry-1.docker.io/bitnamicharts/postgresql

helm install -f helm-config/redis-helm-values.yaml redis bitnami/redis
#helm install -f helm-config/redis-helm-values.yaml stock-redis bitnami/redis
#helm install -f helm-config/redis-helm-values.yaml order-redis bitnami/redis

# helm install kafka-microservice oci://registry-1.docker.io/bitnamicharts/kafka
# helm install kafka-database oci://registry-1.docker.io/bitnamicharts/kafka
