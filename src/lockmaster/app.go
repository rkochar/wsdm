package main

import (
	"log"
	"main/shared"
	"net/http"
)

type Action struct {
	nextMessage string
	topic       string
}

const API_GATEWAY = "http://api-gateway-service-0:27017/"

// Maps incoming message to outgoing message
var successfulActionMap = map[string]Action{
	// Normal checkout
	"START-CHECKOUT-SAGA": {"START-SUBTRACT-STOCK", "stock-syn"},
	"END-SUBTRACT-STOCK":  {"START-MAKE-PAYMENT", "payment-syn"},
	"END-MAKE-PAYMENT":    {"START-UPDATE-ORDER", "order-syn"},
	"END-UPDATE-ORDER":    {"END-CHECKOUT-SAGA", ""},
	// Rollback checkout
	"END-CANCEL-PAYMENT": {"START-READD-STOCK", "stock-syn"},
	"END-READD-STOCK":    {"END-CHECKOUT-SAGA", ""},
}

// Maps message before ABORT to outgoing message
var failActionMap = map[string]Action{
	// Stock Fails
	"START-SUBTRACT-STOCK": {"END-CHECKOUT-SAGA", ""},
	// Payment Fails
	"START-MAKE-PAYMENT": {"START-READD-STOCK", "stock-syn"},
	// Order Fails
	"START-UPDATE-ORDER": {"START-CANCEL-PAYMENT", "payment-syn"},
}

var dbConn MySQLConnection

func main() {
	dbConn = makeMySQLConnection()
	defer dbConn.db.Close()

	shared.SetUpKafkaListener(
		[]string{"order", "stock", "payment"}, true,
		func(message *shared.SagaMessage) (*shared.SagaMessage, string) {

			var nextAction Action
			var messageResponseAvailable bool
			statusCallback := http.StatusOK

			if message.Name == "ABORT-CHECKOUT-SAGA" {
				_, previousLog := dbConn.getLatestSagaLog(message.SagaID)
				_, previousMessage := sagaLogToSagaMessage(previousLog)

				nextAction, messageResponseAvailable = failActionMap[previousMessage.Name]
				statusCallback = routeCheckoutCall(message.Order.OrderID, http.StatusBadRequest)
			} else {
				nextAction, messageResponseAvailable = successfulActionMap[message.Name]
				if nextAction.nextMessage == "END-CHECKOUT-SAGA" {
					statusCallback = routeCheckoutCall(message.Order.OrderID, http.StatusOK)
				}
			}

			log.Print(message.Order.OrderID, http.StatusBadRequest, statusCallback)

			if !messageResponseAvailable {
				return nil, ""
			}

			if message.SagaID == -1 {
				_, sagaID := dbConn.createSaga()
				message.SagaID = *sagaID
			}

			_, sagaLogIn := sagaMessageToSagaLog(message)
			dbConn.insertSagaLog(sagaLogIn)

			outMessage := shared.SagaMessage{
				Name:   nextAction.nextMessage,
				SagaID: message.SagaID,
				Order:  message.Order,
			}

			_, sagaLogOut := sagaMessageToSagaLog(&outMessage)
			dbConn.insertSagaLog(sagaLogOut)

			return &outMessage, nextAction.topic
		},
	)
}

func routeCheckoutCall(orderID string, status int) int {

	backendURL := API_GATEWAY + orderID + "/" + string(status)
	resp, err := http.Get(backendURL)
	if err != nil {
		log.Printf("\nFailed to make service call: %v", err)
		return http.StatusBadRequest
	}
	return resp.StatusCode
}
