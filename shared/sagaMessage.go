package shared

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type SagaMessage struct {
	Name   string
	SagaID string
	Order  Order
}

func parseSagaMessage(message string) (error, *SagaMessage) {
	parts := strings.Split(message, "_")

	order := Order{}

	unmarshalErr := json.Unmarshal([]byte(parts[2]), &order)
	if unmarshalErr != nil {
		return unmarshalErr, nil
	}

	return nil, &SagaMessage{
		Name:   parts[0],
		SagaID: parts[1],
		Order:  order,
	}
}

func sendSagaMessage(message *SagaMessage, conn *kafka.Conn) error {
	jsonByteArray, marshalError := json.Marshal(message.Order)
	if marshalError != nil {
		return marshalError
	}

	messageBuffer := bytes.Buffer{}
	messageBuffer.WriteString("START_CHECKOUT-SAGA_")
	messageBuffer.WriteString(message.SagaID)
	messageBuffer.WriteString("_")
	messageBuffer.Write(jsonByteArray)

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, writeErr := conn.Write(messageBuffer.Bytes())
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func SagaMessageConvertStartToEnd(message *SagaMessage) *SagaMessage {
	parts := strings.Split(message.Name, "-")
	parts[0] = "END"
	return &SagaMessage{
		Name:   strings.Join(parts, "-"),
		SagaID: message.SagaID,
		Order:  message.Order,
	}
}
