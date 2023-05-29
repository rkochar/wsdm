#!/bin/bash

echo "StockDB"
./kubernetes/stockdb/deploy-stockdb.sh

echo "OrderDB"
./kubernetes/orderdb/deploy-orderdb.sh

echo "PaymentDB"
./kubernetes/paymentdb/deploy-paymentdb.sh
