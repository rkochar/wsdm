package main

import (
	"log"

	"WDM-G1/shared"
)

type Action struct {
	nextMessage string
	topic       string
}

// Maps incoming message to outgoing message
var successfulActionMap = map[string]Action{
	// Normal checkout
	"START-CHECKOUT-SAGA": {"START-SUBTRACT-STOCK", "order-syn"},
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

var db MySQLConnection

func main() {
	db = MySQLConnection{}
	defer db.db.Close()
	db.init()

	db.printAllSAGAs()

	shared.SetUpKafkaListener(
		[]string{"order", "stock", "payment"},
		func(message *shared.SagaMessage) (*shared.SagaMessage, string) {

			var nextAction Action
			var messageResponseAvailable bool

			if message.Name == "ABORT-CHECKOUT-SAGA" {
				// TODO: get last successful message name from log
				// nextAction = failActionMap[]

				log.Printf("Abort not supported yet")
				return nil, ""
			} else {
				nextAction, messageResponseAvailable = successfulActionMap[message.Name]
			}

			if !messageResponseAvailable {
				return nil, ""
			}

			// TODO: write incoming message to database

			if message.SagaID == "" {
				// TODO: add new id from database for next message
			}

			outMessage := shared.SagaMessage{
				Name:   nextAction.nextMessage,
				SagaID: message.SagaID,
				Order:  message.Order,
			}

			// TODO: write outgoing message to database

			return &outMessage, nextAction.topic
		},
	)
}
