package models

type OrderState int

const (
	Pending OrderState = iota
	Confirmed
	Processing
	Shipped
	Delivered
	Cancelled
	Returned
)

func (s OrderState) IsStockTaken() bool {
	switch s {
	case Pending:
		return false
	case Confirmed:
		return true
	case Processing:
		return true
	case Shipped:
		return true
	case Delivered:
		return true
	case Cancelled:
		return true
	case Returned:
		return false
	default:
		return false
	}
}

type Order struct {
	ID        uint       `json:"id"`
	ProductID uint       `json:"product_id"`
	Quantity  uint       `json:"quantity"`
	State     OrderState `json:"state"`
}

func (o Order) TakenStock() uint {
	if o.State.IsStockTaken() {
		return o.Quantity
	}
	return 0
}
