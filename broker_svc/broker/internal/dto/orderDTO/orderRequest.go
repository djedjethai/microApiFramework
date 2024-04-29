package orderDTO

import (
	e "gitlab.com/grpasr/common/errors/json"
)

type Order struct {
	ID              string     `json:"id"`
	Items           []LineItem `json:"items,omitempty"`
	ShippingAddress string     `json:"shipping_address"`
}

type LineItem struct {
	ItemCode string `json:"item_code"`
	Quantity int    `json:"quantity"`
}

func (l LineItem) Validate() e.IError {
	if l.Quantity < 0 {
		return e.NewCustomHTTPStatus(e.StatusBadRequest, "", "Quantity must be positive")
	}
	return nil
}
