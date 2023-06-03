package main

import (
	"main/shared"
)

type Action struct {
	nextMessage string
	topic       string
}

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

			if message.Name == "ABORT-CHECKOUT-SAGA" {
				_, previousLog := dbConn.getLatestSagaLog(message.SagaID)
				_, previousMessage := sagaLogToSagaMessage(previousLog)

				nextAction, messageResponseAvailable = failActionMap[previousMessage.Name]
			} else {
				nextAction, messageResponseAvailable = successfulActionMap[message.Name]
			}

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
