#!/bin/bash

echo "Deploying StockDB"
./k8s/mongodb/stockdb/delete-stockdb.sh


echo "Deploying OrderDB"
./k8s/mongodb/orderdb/delete-orderdb.sh

echo "Deploying PaymentDB"
./k8s/mongodb/paymentdb/delete-paymentdb.sh
