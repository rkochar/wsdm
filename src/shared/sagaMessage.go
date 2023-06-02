package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type SagaMessage struct {
	Name   string
	SagaID int64
	Order  Order
}

func ParseSagaMessage(message string) (error, *SagaMessage) {
	parts := strings.SplitN(message, "_", 3)

	if len(parts) < 2 {
		return errors.New("not enough parts to scan"), nil
	}

	order := Order{}

	unmarshalErr := json.Unmarshal([]byte(parts[2]), &order)
	if unmarshalErr != nil {
		fmt.Println("JSON Unmarshal error", unmarshalErr)
		return unmarshalErr, nil
	}
	convErr, sagaIntID := ConvertStringToInt(parts[1])
	if convErr != nil {
		fmt.Println("Convert string to int error:", convErr)
		return convErr, nil
	}

	return nil, &SagaMessage{
		Name:   parts[0],
		SagaID: *sagaIntID,
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
