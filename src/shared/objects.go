package shared

type Order struct {
	OrderID   string   `json:"order_id"`
	Paid      bool     `json:"paid"`
	Items     []string `json:"items"`
	UserID    string   `json:"user_id"`
	TotalCost float64  `json:"total_cost"`
}

type Item struct {
	ItemID string  `json:"item_id"`
	Stock  int64   `json:"stock"`
	Price  float64 `json:"price"`
}

type User struct {
	UserID string  `json:"user_id"`
	Credit float64 `json:"credit"`
}

type Payment struct {
	UserID  string  `json:"user_id"`
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Paid    bool    `json:"paid"`
}
