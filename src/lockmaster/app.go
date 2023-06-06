package main

import (
	"fmt"
	"log"
	"net/http"

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

var numInstances *int64

// var dbConn MySQLConnection
var dbConnections []MySQLConnection

func main() {
	var instanceNumErr error
	instanceNumErr, numInstances = shared.GetNumOfServices(shared.StockService)
	if instanceNumErr != nil {
		log.Fatal(instanceNumErr)
	}
	setupDBConnections()
	for i := 0; i < int(*numInstances); i++ {
		defer dbConnections[i].db.Close()
	}

	shared.SetUpKafkaListener(
		[]string{"order", "stock", "payment"}, true,
		func(message *shared.SagaMessage) (*shared.SagaMessage, string) {

			var nextAction Action
			var messageResponseAvailable bool
			statusCallback := http.StatusOK

			if message.Name == "ABORT-CHECKOUT-SAGA" {
				dbConn := getDBConnection(message.SagaID)
				_, previousLog := dbConn.getLatestSagaLog(message.SagaID)
				_, previousMessage := sagaLogToSagaMessage(previousLog)

				nextAction, messageResponseAvailable = failActionMap[previousMessage.Name]
				statusCallback = shared.RouteCheckoutCall(message.Order.OrderID, http.StatusBadRequest)
			} else {
				nextAction, messageResponseAvailable = successfulActionMap[message.Name]
				if nextAction.nextMessage == "END-CHECKOUT-SAGA" {
					statusCallback = shared.RouteCheckoutCall(message.Order.OrderID, http.StatusOK)
				}
			}

			log.Print(message.Order.OrderID, http.StatusBadRequest, statusCallback)

			if !messageResponseAvailable {
				return nil, ""
			}

			if message.SagaID == -1 {
				dbConn := getDBConnection(message.SagaID)
				_, sagaID := dbConn.createSaga()
				message.SagaID = *sagaID
			}

			_, sagaLogIn := sagaMessageToSagaLog(message)
			dbConnIn := getDBConnection(message.SagaID)
			dbConnIn.insertSagaLog(sagaLogIn)

			outMessage := shared.SagaMessage{
				Name:   nextAction.nextMessage,
				SagaID: message.SagaID,
				Order:  message.Order,
			}

			_, sagaLogOut := sagaMessageToSagaLog(&outMessage)
			dbConnOut := getDBConnection(outMessage.SagaID)
			dbConnOut.insertSagaLog(sagaLogOut)

			return &outMessage, nextAction.topic
		},
	)
}

func setupDBConnections() {
	dbConnections := make([]MySQLConnection, *numInstances)

	for i := 0; i < int(*numInstances); i++ {
		fmt.Println("Setting up DB connection %d\n", i)

		dbConnections[i] = makeMySQLConnection(i)
	}
}

func getDBConnection(sagaID int64) MySQLConnection {
	databaseNum := sagaID % *numInstances
	return dbConnections[databaseNum]
}
