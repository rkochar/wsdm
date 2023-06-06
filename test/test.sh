echo "Creating user id"
USER_ID=$(curl localhost:8080/payment/create_user | jq '.user_id')
echo "[USER_ID]: $USER_ID\n"

echo "giving user money"
curl localhost:8080/payment/add_funds/$USER_ID/69420
echo "\n"

echo "Creating stock"
ITEM_ID=$(curl localhost:6969/stock/item/create/42 | jq '.item_id')
echo "[ITEM_ID]: $ITEM_ID\n"

echo "Adding 100 items to stock_id"
curl localhost:6969/stock/add/$ITEM_ID/100
echo "\n"

echo "Creating order"
echo "$(curl localhost:6969/orders/create/$USER_ID)"
ORDER_ID=$(curl localhost:6969/orders/create/$USER_ID | jq '.order_id')
echo "[ORDER_ID]: $ORDER_ID\n"

echo "Adding item to order"
curl localhost:6969/orders/addItem/$ORDER_ID/$ITEM_ID

echo "Checking out"
curl localhost:6969/orders/checkout/$ORDER_ID
