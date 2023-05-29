#!/bin/bash

echo "StockDB"
./kubernetes/stockdb/delete-stockdb.sh

echo "OrderDB"
./kubernetes/orderdb/delete-orderdb.sh

echo "PaymentDB"
./kubernetes/paymentdb/delete-paymentdb.sh

