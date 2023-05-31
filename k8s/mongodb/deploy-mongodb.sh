#!/bin/bash

echo "Deploying StockDB"
./k8s/mongodb/stockdb/deploy-stockdb.sh


echo "Deploying OrderDB"
./k8s/mongodb/orderdb/deploy-orderdb.sh

echo "Deploying PaymentDB"
./k8s/mongodb/paymentdb/deploy-paymentdb.sh
