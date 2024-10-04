package handlers

import (
	"TouchySarun/chp_order_backend/internal/services"
	"fmt"
	"net/http"
)

func CreateSkus (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	skus, err := services.ReadSkus("skus_sample.csv")
	if err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Error:%v", err), http.StatusInternalServerError)
		return
	}

	for _, sku := range skus {
		fmt.Printf("Name: %s, Goods: %v\n", sku.Name, sku.Goods)
		services.CreateSku(ctx, sku)
	}
	services.WriteResponseSuccess(&w, "Success get sku from csv file")
}