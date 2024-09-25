package handlers

import (
	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func GetSku(w http.ResponseWriter, r *http.Request) {
	barcode := mux.Vars(r)["barcode"]

	ctx := r.Context()
	skuData, skuErr := services.GetSkuByBarcode(ctx, barcode)
	if skuErr != nil || skuData == nil || skuData.Id == nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Product not found. %v", skuErr), http.StatusNotFound)
		return
	}
	services.WriteResponseSuccess(&w, skuData)
}
func GetCreateOrderData(w http.ResponseWriter, r *http.Request) {
	barcode := mux.Vars(r)["barcode"]
	branch := mux.Vars(r)["branch"]


	ctx := r.Context()
	skuData, skuErr := services.GetSkuByBarcode(ctx, barcode)
	if skuErr != nil || skuData == nil || skuData.Id == nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Product not found. %v", skuErr), http.StatusNotFound)
		return
	}
	skuId := *skuData.Id
	orderData, orderErr := services.GetLatestOrder(ctx, skuId, branch)
	if orderErr != nil{
		services.WriteResponseErr(&w, fmt.Sprintf("Failed, getting latest order. %v", orderErr), http.StatusNotFound)
		return
	}
	latestDate, latestErr := services.GetLatestSuccessOrderDate(ctx, skuId, branch)
	if latestErr != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Failed, getting latest success order's date. %v", latestErr), http.StatusNotFound)
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
	services.WriteResponseSuccess(&w, res)
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.OrderCreateReqeust
	ctx := r.Context()

	// Decode the JSON request body into the req struct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		services.WriteResponseErr(&w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request fields (check if any required fields are missing)
	if req.Branch == "" || req.Name == "" || req.UtqName == "" || req.UtqQty == 0 ||
		req.Code == "" || req.Sku == "" || req.Ap == "" || req.Qty == 0 ||
		req.Cat == "" || req.Bnd == "" || req.CreBy == "" {
		services.WriteResponseErr(&w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create an order object from the request data
	order := models.Order{
		Branch:    req.Branch,
		Name:      req.Name,
		UtqName:   req.UtqName,
		UtqQty:    req.UtqQty,
		Code:      req.Code,
		Sku:       req.Sku,
		Ap:        req.Ap,
		Qty:       req.Qty,
		LeftQty:   req.Qty,
		Cat:       req.Cat,
		Bnd:       req.Bnd,
		CreBy:     req.CreBy,
		StartDate: time.Now(),
		Status:    "init",
	}

	orderHistory := models.OrderHistory{
		Status: "init",
		Date: time.Now(),
		CreBy: req.CreBy,
	}

	// Call the service to create the order
	id, err := services.CreateOrder(ctx, order)
	if err != nil || id == nil {
		services.WriteResponseErr(&w, "Failed to create order", http.StatusInternalServerError)
		return
	}
	ohid, err := services.CreateOrderHistory(ctx, *id, orderHistory)
	if err != nil || ohid == nil {
		services.WriteResponseErr(&w, "Failed to create orderhistory", http.StatusInternalServerError)
	}
	// Write a success response
	services.WriteResponseSuccess(&w, id)
}