package main

import (
	"WDM-G1/shared"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var messageTypeMapStringToInt = map[string]int64{
	"START": 1,
	"END":   2,
	"ABORT": 3,
}

var messageEventMapStringToInt = map[string]int64{
	"MAKE-PAYMENT":   1,
	"CANCEL-PAYMENT": 2,
	"CHECKOUT-SAGA":  3,
	"CANCEL-SAGA":    4,
	"SUBTRACT-STOCK": 5,
	"READD-STOCK":    6,
	"UPDATE-ORDER":   7,
}

var messageTypeMapIntToString = map[int64]string{
	1: "START",
	2: "END",
	3: "ABORT",
}

var messageEventMapIntToString = map[int64]string{
	1: "MAKE-PAYMENT",
	2: "CANCEL-PAYMENT",
	3: "CHECKOUT-SAGA",
	4: "CANCEL-SAGA",
	5: "SUBTRACT-STOCK",
	6: "READD-STOCK",
	7: "UPDATE-ORDER",
}

func sagaMessageToSagaLog(sagaMessage *shared.SagaMessage) (error, *SagaLog) {
	msgTypErr, messageType := getMessageTypeInt(sagaMessage.Name)
	if msgTypErr != nil {
		return msgTypErr, nil
	}
	fmt.Printf("Message Type: %d", messageType)

	msgEventErr, messageEvent := getMessageEventInt(sagaMessage.Name)
	if msgEventErr != nil {
		return msgEventErr, nil
	}
	fmt.Printf("Message Event: %d", messageEvent)

	return nil, &SagaLog{
		SagaID:       sagaMessage.SagaID,
		MessageType:  messageType,
		MessageEvent: messageEvent,
	}
}

func sagaLogToSagaMessage(sagaLog *SagaLog) (error, *shared.SagaMessage) {
	var order shared.Order
	unMarshallErr := json.Unmarshal([]byte(sagaLog.SagaContents), &order)
	if unMarshallErr != nil {
		return unMarshallErr, nil
	}

	typeConvErr, messageType := getMessageTypeString(sagaLog.MessageType)
	if typeConvErr != nil {
		return typeConvErr, nil
	}
	nameConvErr, messageName := getMessageEventString(sagaLog.MessageEvent)
	if nameConvErr != nil {
		return nameConvErr, nil
	}

	name := strings.Join([]string{messageType, messageName}, "-")

	message := shared.SagaMessage{
		Name:   name,
		SagaID: sagaLog.SagaID,
		Order:  order,
	}

	return nil, &message
}

func getMessageTypeInt(messageName string) (error, int64) {
	parts := strings.Split(messageName, "-")

	messageType, found := messageTypeMapStringToInt[parts[0]]
	if !found {
		errorMsg := fmt.Sprintf("invalid message type: %s", parts[0])
		return errors.New(errorMsg), -1
	}
	return nil, messageType
}

func getMessageEventInt(messageName string) (error, int64) {
	parts := strings.Split(messageName, "-")

	messageEventStr := strings.Join(parts[1:2], "-")
	messageEvent, found := messageEventMapStringToInt[messageEventStr]
	if !found {
		errorMsg := fmt.Sprintf("invalid message event: %s", messageEventStr)
		return errors.New(errorMsg), -1
	}
	return nil, messageEvent
}

func getMessageTypeString(messageCode int64) (error, string) {
	messageType, found := messageTypeMapIntToString[messageCode]
	if !found {
		errorMsg := fmt.Sprintf("invalid message type: %d", messageCode)
		return errors.New(errorMsg), ""
	}
	return nil, messageType
}

func getMessageEventString(messageCode int64) (error, string) {
	messageEvent, found := messageEventMapIntToString[messageCode]
	if !found {
		errorMsg := fmt.Sprintf("invalid message event: %d", messageCode)
		return errors.New(errorMsg), ""
	}
	return nil, messageEvent
}
