package handlers

import (
	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func GetSku(w http.ResponseWriter, r *http.Request) {
	barcode := mux.Vars(r)["barcode"]

	ctx := r.Context()
	skuData, skuErr := services.GetSkuByBarcode(ctx, barcode)
	if skuErr != nil || skuData == nil || skuData.Id == nil {
		http.Error(w, "Product not found.", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(skuData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
func GetCreateOrderData(w http.ResponseWriter, r *http.Request) {
	barcode := mux.Vars(r)["barcode"]
	branch := mux.Vars(r)["branch"]


	ctx := r.Context()
	skuData, skuErr := services.GetSkuByBarcode(ctx, barcode)
	if skuErr != nil || skuData == nil || skuData.Id == nil {
		http.Error(w, "Product not found.", http.StatusNotFound)
		return
	}
	skuId := *skuData.Id
	orderData, orderErr := services.GetLatestOrder(ctx, skuId, branch)
	if orderErr != nil{
		http.Error(w, "Failed, getting latest order", http.StatusNotFound)
		return
	}
	latestDate, latestErr := services.GetLatestSuccessOrderDate(ctx, skuId, branch)
	if latestErr != nil {
		http.Error(w, "Failed, getting latest success order's date", http.StatusNotFound)
		return
	}

	res := models.OrderCreateData{
		Sku: *skuData,
	}
	if orderData != nil {
		res.Order = *orderData
	}
	if latestDate != nil {
		res.LastOrderDate = *latestDate
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {

}