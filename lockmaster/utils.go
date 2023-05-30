package main

import (
	"WDM-G1/shared"
	"errors"
	"fmt"
	"strings"
)

var messageTypeMap = map[string]int64{
	"START": 1,
	"END":   2,
	"ABORT": 3,
}

var messageEventMap = map[string]int64{
	"MAKE-PAYMENT":   1,
	"CANCEL-PAYMENT": 2,
	"CHECKOUT-SAGA":  3,
	"CANCEL-SAGA":    4,
	"SUBTRACT-STOCK": 5,
	"READD-STOCK":    6,
	"UPDATE-ORDER":   7,
}

func MessageToSagaLog(sagaMessage *shared.SagaMessage) (error, *SagaLog) {
	msgTypErr, messageType := GetMessageType(sagaMessage.Name)
	if msgTypErr != nil {
		return msgTypErr, nil
	}
	fmt.Printf("Message Type: %d", messageType)

	msgEventErr, messageEvent := GetMessageEvent(sagaMessage.Name)
	if msgEventErr != nil {
		return msgEventErr, nil
	}
	fmt.Printf("Message Event: %d", messageEvent)

	intConvErr, intID :=

	return nil, &SagaLog{
		SagaID: sagaMessage.SagaID,
		MessageType: messageType,
		MessageEvent: messageEvent,
	}
}

func GetMessageType(messageName string) (error, int64) {
	parts := strings.Split(messageName, "-")

	messageType, found := messageTypeMap[parts[0]]
	if !found {
		errorMsg := fmt.Sprintf("invalid message type: %s", parts[0])
		return errors.New(errorMsg), -1
	}
	return nil, messageType
}

func GetMessageEvent(messageName string) (error, int64) {
	parts := strings.Split(messageName, "-")

	messageEventStr := strings.Join(parts[1:2], "-")
	messageEvent, found := messageEventMap[messageEventStr]
	if !found {
		errorMsg := fmt.Sprintf("invalid message event: %s", messageEventStr)
		return errors.New(errorMsg), -1
	}
	return nil, messageEvent
}
