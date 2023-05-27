package main

import (
	"encoding/json"
	"strings"
)

type Order struct {
	OrderID   string   `json:"order_id"`
	Paid      bool     `json:"paid"`
	Items     []string `json:"items"`
	UserID    string   `json:"user_id"`
	TotalCost float64  `json:"total_cost"`
}

type SagaMessage struct {
	name   string
	sagaID string
	order  Order
}

func parseSagaMessage(message string) (error, *SagaMessage) {
	parts := strings.Split(message, "_")

	order := Order{}

	unmarshalErr := json.Unmarshal([]byte(parts[2]), &order)
	if unmarshalErr != nil {
		return unmarshalErr, nil
	}

	return nil, &SagaMessage{
		name:   parts[0],
		sagaID: parts[1],
		order:  order,
	}
}
