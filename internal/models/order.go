package models

import (
	"time"
)

type OrderCreateData struct {
	Sku Sku
	Order Order
	LastOrderDate string
}

type Sku struct {
	Id *string
	Name string	`firestore:"name"`
	Ap string	`firestore:"ap"`
	Img string	`firestore:"img"`
	Cat string	`firestore:"cat"`
	Bnd string	`firestore:"bnd"`
	Goods []Goods	`firestore:"goods"`
}
type Goods struct {
	Code string	`firestore:"code"`
	UtqName string	`firestore:"utqName"`
	UtqQty string	`firestore:"utqQty"`
	Price0 string	`firestore:"price0"`
	Prict8 string	`firestore:"prict8"`
}
type Order struct {
	Id	*string
	Branch string	`firestore:"branch"`
	Name string	`firestore:"name"`
	UtqName string	`firestore:"utqName"`
	UtqQty int	`firestore:"utqQty"`
	Code string	`firestore:"code"`
	Sku string	`firestore:"sku"`
	Ap string	`firestore:"ap"`
	Qty int	`firestore:"qty"`
	LeftQty int	`firestore:"leftQty"`
	Cat string	`firestore:"cat"`
	Bnd string	`firestore:"bnd"`
	CreBy string	`firestore:"creBy"`
	StartDate time.Time	`firestore:"startDate"`
	EndDate 	*time.Time `firestore:"endDate"`
	Status string	`firestore:"status"`
	LstUpd time.Time `firestore:"lstUpd"`
	History *[]OrderHistory `firestore:"history"`
}
type OrderHistory struct {
	Status string `firestore:"status"`
	Date time.Time `firestore:"date"`
	CreBy string `firestore:"creBy"`
	OldUtqName		*string `firestore:"oldUtqName"`
	NewUtqName		*string `firestore:"newUtqName"`
	OldQty			*int `firestore:"oldQty"`
	NewQty		*int `firestore:"newQty"`
	Remark		*string `firestore:"remark"`
}

type OrderCreateReqeust struct {
	Branch string `json:"branch"`
	Name string `json:"name"`
	UtqName string `json:"utqName"`
	UtqQty int `json:"utqQty"`
	Code string `json:"code"`
	Sku string `json:"sku"`
	Ap string `json:"ap"`
	Qty int `json:"qty"`
	Cat string `json:"cat"`
	Bnd string `json:"bnd"`
	CreBy string `json:"creBy"`
}

type OrderEditRequest struct {
	Id string `json:"id"`
	UtqName string `json:"utqName"`
	UtqQty int `json:"utqQty"`
	Code string `json:"code"`
	Qty int `json:"qty"`
	CreBy string `json:"creBy"`
}