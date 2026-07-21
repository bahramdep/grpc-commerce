package inventory

import "time"

type Status int

const (
	StatusUnspecified Status = iota
	StatusReserved
	StatusReleased
)

type Item struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type Reservation struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Items     []Item    `json:"items"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
