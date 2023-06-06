package shared

import "github.com/google/uuid"

type Order struct {
	ID        uuid.UUID `bson:"_id"`
	OrderID   string    `json:"order_id"`
	Paid      bool      `json:"paid"`
	Items     []string  `json:"items"`
	UserID    string    `json:"user_id"`
	TotalCost int64     `json:"total_cost"`
}

type Item struct {
	ID     uuid.UUID `bson:"_id"`
	ItemID string    `json:"item_id"`
	Stock  int64     `json:"stock"`
	Price  int64     `json:"price"`
}

type User struct {
	ID     uuid.UUID `bson:"_id"`
	UserID string    `json:"user_id"`
	Credit int64     `json:"credit"`
}

type Payment struct {
	ID      uuid.UUID `bson:"_id"`
	UserID  string    `json:"user_id"`
	OrderID string    `json:"order_id"`
	Amount  int64     `json:"amount"`
	Paid    bool      `json:"paid"`
}
