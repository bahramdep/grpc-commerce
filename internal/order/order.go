package order

import "time"


type Status int

const (
	StatusUnspecified Status = iota
	StatusPending
)

type Item struct{
	ProductID string
	Quantity int32
}


type Order struct {
	ID string
	CustomerID string
	Items []Item
	Status Status
	CreatedAt time.Time
}