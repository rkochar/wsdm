#!/bin/bash

echo "StockDB"
./mongo/stockdb/deploy-stockdb.sh

echo "OrderDB"
./mongo/orderdb/deploy-orderdb.sh

echo "PaymentDB"
./mongo/paymentdb/deploy-paymentdb.sh
