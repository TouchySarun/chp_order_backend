package handlers

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	defFirestore "cloud.google.com/go/firestore"

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
	id, cerr := services.CreateOrder(ctx, order)
	if cerr != nil || id == nil {
		services.WriteResponseErr(&w, "Failed to create order", http.StatusInternalServerError)
		return
	}
	coherr := services.CreateOrderHistory(ctx, *id, orderHistory)
	if coherr != nil  {
		services.WriteResponseErr(&w, "Failed to create orderhistory", http.StatusInternalServerError)
	}
	// Write a success response
	services.WriteResponseSuccess(&w, id)
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
	fmt.Println("Success decoder")
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
	fmt.Println("Success validate")
	// Create update fields for Firestore
	updatedFields, mufErr := makeOrderUpdateField(req)
	if mufErr != nil {
		services.WriteResponseErr(&w, "No fields to update", http.StatusBadRequest)
		return
	}
	fmt.Printf("Success make order update field. :%v \n",updatedFields)
	// Start Firestore transaction to ensure consistency
	err = firestore.Client.RunTransaction(ctx, func(ctx context.Context, tx *defFirestore.Transaction) error {
		fmt.Printf("order id : %v \n",req.Id)
		// Edit order within the transaction
		editErr := services.EditOrder(ctx, req.Id, *updatedFields)
		if editErr != nil {
			return editErr
		}
		fmt.Println("Success edit order")
		// Create order history within the transaction
		historyErr := services.CreateOrderHistory(ctx, req.Id, makeOrderHistoryUpdateField(req,*order))
		if historyErr != nil {
			return fmt.Errorf("failed to create order history: %v", historyErr)
		}
		fmt.Println("Success create history")
		return nil
	})
	// Handle any transaction errors
	if err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Transaction failed: %v", err), http.StatusInternalServerError)
		return
	}
}

func makeOrderHistoryUpdateField (req models.OrderEditRequest, order models.Order) (models.OrderHistory){
	var newQty = order.Qty + req.Qty
	var newUtqName = order.UtqName
	if req.UtqName != "" {
		newUtqName = req.UtqName
	}
	orderHistory := models.OrderHistory{
		Status:     "edit",
		Date:       time.Now(),
		CreBy:      req.CreBy,
		OldQty:     &order.Qty,
		OldUtqName: &order.UtqName,
		NewQty:     &newQty,
		NewUtqName: &newUtqName,
	}
	return orderHistory
}

func makeOrderUpdateField (req models.OrderEditRequest) (*map[string]interface{}, error) {
	updatedFields := make(map[string]interface{})
	reqValue := reflect.ValueOf(req) // get array of value from req
	reqType := reflect.TypeOf(req) // get array of type from req

	if reqValue.NumField() == 0 {
		return nil, fmt.Errorf("no input fields")
	}

	for i:=0; i< reqValue.NumField(); i++ {
		v := reqValue.Field(i) // get req[i].value
		t	:= reqType.Field(i) // get req[i].type
		n := t.Tag.Get("json") // get type name

		// if have value add to update fields
		switch v.Kind() {
			case reflect.String: // Handle strings (value type)
				if v.String() != "" { // Check if string is not empty
					updatedFields[n] = v.String()
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // Handle integer types
				if v.Int() != 0 { // Check if integer is non-zero
					updatedFields[n] = v.Int()
				}
		}
	}

	if len(updatedFields) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	if updatedFields["qty"] != nil {
		updatedFields["leftQty"] = req.Qty
	}
	updatedFields["lstUpd"] = time.Now()
	return &updatedFields, nil
}