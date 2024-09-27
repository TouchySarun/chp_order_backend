package models

type Shipping struct {
	Id string `json:"id"`
	OrderId string `json:"orderId"`
	Qty int `json:"qty"`
	Branch string `json:"branch"`
}

type ConfirmShippingRequest struct {
	Branch string `json:"branch"`
	CreBy string	`json:"creBy"`
}