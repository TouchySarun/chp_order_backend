package models

type Shipping struct {
	Id string `json:"id"`
	OrderId string `json:"orderId"`
	Qty int `json:"qty"`
	Branch int `json:"branch"`
}