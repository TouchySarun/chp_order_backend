package handlers

import (
	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	dcerr := json.NewDecoder(r.Body).Decode(&req)
	if dcerr != nil {
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
	services.CreateOrderAndOrderHistory(ctx, order, orderHistory)
	// Write a success response
	services.WriteResponseSuccess(&w, order)
}

func EditOrder(w http.ResponseWriter, r *http.Request) { 
	var req models.OrderEditRequest
	ctx := r.Context()
	// Decode the JSON request body into the req struct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		services.WriteResponseErr(&w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Validate the request fields (check if required fields are missing)
	if req.Id == "" || req.CreBy == "" {
		services.WriteResponseErr(&w, "Missing required fields", http.StatusBadRequest)
		return
	}
	// Get the current order to update
	order, orderErr := services.GetOrder(ctx, req.Id)
	if orderErr != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("failed to retrieve order: %v", orderErr), http.StatusBadRequest)
		return 
	}
	// Check that either newQty or newUtqQty is provided
	if req.Qty == 0 && req.Code == order.Code {
		services.WriteResponseErr(&w, "Missing both Qty and UtqQty fields", http.StatusBadRequest)
		return
	}
	// Create update fields for Firestore
	updatedFields, mufErr := services.MakeOrderUpdateField(req)
	if mufErr != nil {
		services.WriteResponseErr(&w, "No fields to update", http.StatusBadRequest)
		return
	}
	// Start Firestore transaction to ensure consistency
	err = services.EditAndCreateOrderHistory(ctx, *order.Id, *updatedFields, services.MakeOrderHistoryUpdateField(req,*order))
	// Handle any transaction errors
	if err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Transaction failed: %v", err), http.StatusInternalServerError)
		return
	}
	services.WriteResponseSuccess(&w, "Success")
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	// Required query parameters
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")
	ctx := r.Context()
	// Convert 'limit' and 'page' to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		http.Error(w, "Invalid or missing 'limit' parameter", http.StatusBadRequest)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		http.Error(w, "Invalid or missing 'page' parameter", http.StatusBadRequest)
		return
	}
	fieldConditions := map[string]string{
		"status": r.URL.Query().Get("status"),
		"creBy":  r.URL.Query().Get("creBy"),
		"ap":     r.URL.Query().Get("ap"),
		"rack":   r.URL.Query().Get("rack"),
		"branch": r.URL.Query().Get("code"),
	}
	code := r.URL.Query().Get("code")
	orders, err := services.GetOrders(ctx, fieldConditions, code,limit, page)
	if err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Failed, Getting orders, %v",err),http.StatusInternalServerError)
	}
	services.WriteResponseSuccess(&w,orders)
}

func UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	var updatedFields = map[string]interface{}{"lstUpd":time.Now()}
	var req models.OrderUpdateStatusRequest

	dcerr := json.NewDecoder(r.Body).Decode(&req)
	if dcerr != nil {
		services.WriteResponseErr(&w, "can't decode request.", http.StatusBadRequest)
		return
	}
	order, oerr := services.GetOrder(ctx, id)
	if oerr != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("can't find order. \n %v\n",oerr), http.StatusBadRequest)
		return
	}
	var oh = models.OrderHistory{
		Date:       time.Now(),
		CreBy:      req.CreBy,
	}
	// create updateFields and orderHistory
	switch req.Status{
	case "picking":{
		// pc click order => just update status, lstUpd ## except status == "shipping"
		if order.Status == "shipping" || order.Status == "picking" {
			services.WriteResponseSuccess(&w, fmt.Sprintf("Must update status to picking (PC click order) but order status is %v, so I do nothing.", order.Status))
			return
		} else {
			updatedFields["status"]="picking"
			oh.Status = "picking"
			break
		}
	}
	case "shipping":{
		// pc click picking done => check if (qty == 0) just update status to left else update status to shipping left qty
		if order.Status != "shipping" && order.Status != "picking" {
			services.WriteResponseSuccess(&w, fmt.Sprintf("Must update status to shipping (PC done picking, order status shuld be picking or shipping) but order status is %v.", order.Status))
			return
		}
		if req.Qty == 0 {
			services.WriteResponseSuccess(&w, "Must update status to shipping and qty is 0, so I do nothing.")
			return
		} else {
			updatedFields["status"] = "shipping"
			updatedFields["leftQty"] = -1 * req.Qty

			newQty := order.LeftQty - req.Qty
			oh.Status = "shipping"
			oh.OldQty = &order.LeftQty
			oh.NewQty = &newQty
			break
		}
	}
	case "done": {
		// br click shipping done => check if (leftQty == 0) 
		if order.LeftQty == 0 {
			updatedFields["status"] = "done"
			updatedFields["endDate"] = time.Now()
			oh.Status = "done"
			break
		} else {
			updatedFields["status"] = "left"
			oh.Status = "left"
			break
		}
	}
	}
	if err := services.EditAndCreateOrderHistory(ctx,id,updatedFields,oh); err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Transaction failed: %v", err), http.StatusInternalServerError)
	} else {
		services.WriteResponseSuccess(&w, "Success")
	}
}