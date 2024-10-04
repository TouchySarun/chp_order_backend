package models

import (
	"time"
)

type OrderCreateData struct {
	Sku Sku `json:"sku"`
	Order *Order `json:"order,omitempty"`
	LastOrderDate *time.Time `json:"lstSuccess,omitempty"`
}

type Sku struct {
	Id *string `json:"id"`
	Name string	`firestore:"name" json:"name"`
	Ap string	`firestore:"ap" json:"ap"`
	Img string	`firestore:"img" json:"img"`
	Cat string	`firestore:"cat" json:"cat"`
	Bnd string	`firestore:"bnd" json:"bnd"`
	Barcodes []string `firestore:"barcodes" json:"barcodes"`
	Goods []Goods	`firestore:"goods" json:"goods"`
}
type Goods struct {
	Code string	`firestore:"code" json:"code"`
	UtqName string	`firestore:"utqName" json:"utqName"`
	UtqQty int	`firestore:"utqQty" json:"utqQty"`
	Price0 string	`firestore:"price0" json:"price0"`
	Prict8 string	`firestore:"prict8" json:"price8"`
}
type Order struct {
	Id	*string
	Branch string	`firestore:"branch" json:"branch"`
	Name string	`firestore:"name" json:"name"`
	UtqName string	`firestore:"utqName" json:"utqName"`
	UtqQty int	`firestore:"utqQty" json:"utqQty"`
	Code string	`firestore:"code" json:"code"`
	Sku string	`firestore:"sku" json:"sku"`
	Ap string	`firestore:"ap" json:"ap"`
	Qty int	`firestore:"qty" json:"qty"`
	LeftQty int	`firestore:"leftQty" json:"leftQty"`
	Cat string	`firestore:"cat" json:"cat"`
	Bnd string	`firestore:"bnd" json:"bnd"`
	CreBy string	`firestore:"creBy" json:"creBy"`
	StartDate time.Time	`firestore:"startDate" json:"startDate"`
	EndDate 	*time.Time `firestore:"endDate" json:"endDate,omitempty"`
	Status string	`firestore:"status" json:"status"`
	LstUpd *time.Time `firestore:"lstUpd" json:"lstUpd,omitempty"`
	History *[]OrderHistory `firestore:"history" json:"history,omitempty"`
}
type OrderHistory struct {
	Status string `firestore:"status" json:"status"`
	Date time.Time `firestore:"date" json:"date"`
	CreBy string `firestore:"creBy" json:"creBy"`
	OldUtqName		*string `firestore:"oldUtqName" json:"oldUtqName"`
	NewUtqName		*string `firestore:"newUtqName" json:"newUtqName"`
	OldQty			*int `firestore:"oldQty" json:"oldQty"`
	NewQty		*int `firestore:"newQty" json:"newQty"`
	Remark		*string `firestore:"remark" json:"remark"`
}

type OrderCreateReqeust struct {
	Ap string `json:"ap"`
	Bnd string `json:"bnd"`
	Branch string `json:"branch"`
	Cat string `json:"cat"`
	Code string `json:"code"`
	CreBy string `json:"creBy"`
	Name string `json:"name"`
	Qty int `json:"qty"`
	Sku string `json:"sku"`
	UtqName string `json:"utqName"`
	UtqQty int `json:"utqQty"`
}

type OrderEditRequest struct {
	Id string `json:"id"`
	UtqName string `json:"utqName"`
	UtqQty int `json:"utqQty"`
	Code string `json:"code"`
	Qty int `json:"qty"`
	CreBy string `json:"creBy"`
}

type OrderUpdateStatusRequest struct {
	Status string `json:"status"`
	CreBy string `json:"creBy"`
	Qty int `json:"qty"`
}