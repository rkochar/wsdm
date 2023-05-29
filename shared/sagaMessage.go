package shared

import (
	"encoding/json"
	"strings"
)

type SagaMessage struct {
	Name   string
	SagaID string
	Order  Order
}

func ParseSagaMessage(message string) (error, *SagaMessage) {
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

func SagaMessageConvertStartToEnd(message *SagaMessage) *SagaMessage {
	parts := strings.Split(message.Name, "-")
	parts[0] = "END"
	return &SagaMessage{
		Name:   strings.Join(parts, "-"),
		SagaID: message.SagaID,
		Order:  message.Order,
	}
}
