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
	parts1 := strings.Split(message, "{")
	jsonPart := "{" + parts1[1]

	parts := strings.Split(parts1[0], "_")

	//parts := strings.Split(message, "_")
	order := Order{}

	if len(parts) < 2 {
		return errors.New("not enough parts to scan!"), nil
	}
	//TODO: crashes because splitting on _ caused the JSON to break
	//TODO: use JSON object
	fmt.Println("Parts:", parts)
	fmt.Println("JSON:", jsonPart)

	unmarshalErr := json.Unmarshal([]byte(jsonPart), &order)
	if unmarshalErr != nil {
		fmt.Println("JSON Unmarshale error", unmarshalErr)
		return unmarshalErr, nil
	}
	convErr, sagaIntID := ConvertStringToInt(parts[2])
	if convErr != nil {
		fmt.Println("Convert string to int error:", convErr)
		return convErr, nil
	}

	return nil, &SagaMessage{
		Name:   parts[0] + "_" + parts[1],
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
