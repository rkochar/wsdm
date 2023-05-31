#!/bin/bash

echo "Deleting mongo"
echo "StockDB"
./kubernetes/stockdb/delete-stockdb.sh

echo "OrderDB"
./kubernetes/orderdb/delete-orderdb.sh

echo "PaymentDB"
./kubernetes/paymentdb/delete-paymentdb.sh

./k8s/delete-microservice.sh


./kafka/delete-kakfa.sh
