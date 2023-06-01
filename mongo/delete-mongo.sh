#!/bin/bash

echo "StockDB"
./mongo/stockdb/delete-stockdb.sh

echo "OrderDB"
./mongo/orderdb/delete-orderdb.sh

echo "PaymentDB"
./mongo/paymentdb/delete-paymentdb.sh

