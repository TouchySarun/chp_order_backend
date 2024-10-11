package handlers

import (
	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"
	"context"
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
		services.WriteResponseErr(&w, skuErr.Error(), http.StatusNotFound)
		return
	}
	// fmt.Printf("success get sku %v \n ", skuData)
	skuId := *skuData.Id
	orderData, orderErr := services.GetLatestOrder(ctx, skuId, branch)
	if orderErr != nil{
		services.WriteResponseErr(&w, orderErr.Error(), http.StatusInternalServerError)
		return
	}
	latestDate, latestErr := services.GetLatestSuccessOrderDate(ctx, skuId, branch)
	if latestErr != nil {
		services.WriteResponseErr(&w, latestErr.Error(), http.StatusInternalServerError)
		return
	}
	res := models.OrderCreateData{
		Sku: *skuData,
	}
	if orderData != nil {
		orderData.History = nil 
		res.Order = orderData
	}
	if latestDate != nil {
		res.LastOrderDate = latestDate
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
func validateGetOrdersInput(ctx context.Context, r *http.Request) (models.OrderQuery, int, error) {
	
	query := models.OrderQuery{
		CreBy		: r.URL.Query().Get("creBy"),
		Rack		: r.URL.Query().Get("rack"),
		Status 	: r.URL.Query()["status"], // Parse as []string
		Ap 			: r.URL.Query().Get("ap"),
		Bnd 		:	r.URL.Query().Get("bnd"),
		Search	: r.URL.Query().Get("search"),
		OrderBy	: r.URL.Query().Get("orderBy"),
	}
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")
	if query.OrderBy == "" {
		return query, http.StatusBadRequest, fmt.Errorf("missing 'order by' parameter")
	}
	// Convert 'limit' and 'page' to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		query.Limit = limit
		return query, http.StatusBadRequest, fmt.Errorf("invalid or missing 'limit' parameter")
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		query.Offset = (page-1)*limit
		return query, http.StatusBadRequest, fmt.Errorf("invalid or missing 'page' parameter")
	}
	// check branch if no branch input set to get all branches
	branch := r.URL.Query()["branch"];
	if len(branch) == 0 {
		b, _ := services.GetBranches(ctx)
		branch = *b
	}
	query.Branch = branch
	// check username, get aps
	aps := make([]string, 0)
	username := r.URL.Query().Get("username")
	if username != "" {
		user, err := services.GetUserByUsername(ctx, username)
		if err != nil {
			return query, http.StatusInternalServerError, err
		}
		rack := user.Rack
		resAps, err := services.GetApsByRack(ctx, rack)
		if err != nil {
			return query, http.StatusInternalServerError, err
		}
		for _, ap := range(*resAps) {
			aps = append(aps, ap.Code)
		}
	}
	if len(aps)>0{
		query.ApContain = aps
	}
	// check code, get codes
	code := r.URL.Query().Get("code")
	if code != "" {
		sku, err := services.GetSkuByBarcode(ctx, code)
		if err != nil {
			return query, http.StatusInternalServerError, err
		}
		query.Code = sku.Barcodes
	}
	return query, http.StatusOK, nil
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	// TODO: return count all(no filter, with filter)
	ctx := r.Context()
	query, status, err := validateGetOrdersInput(ctx, r)
	if err != nil {
		services.WriteResponseErr(&w, err.Error(), status)
		return;
	}
	orders, err := services.GetOrders(ctx, query)
	if err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Failed, Getting orders, %v",err),http.StatusInternalServerError)
		return;
	}
	services.WriteResponseSuccess(&w,orders)
}

func UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	var req models.OrderUpdateStatusRequest
	
	dcerr := json.NewDecoder(r.Body).Decode(&req)
	if dcerr != nil {
		services.WriteResponseErr(&w, "can't decode request.", http.StatusBadRequest)
		return
	}
	message, err := services.UpdateStatus(ctx, id, req.Status, req.Qty, req.CreBy)
	if err != nil {
		services.WriteResponseErr(&w, err.Error(), http.StatusInternalServerError)
	} else {
		services.WriteResponseSuccess(&w, message)
	}
}