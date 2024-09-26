package handlers

import (
	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreateShipping (w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	var req models.Shipping
	dcerr := json.NewDecoder(r.Body).Decode(&req)
	if dcerr != nil {
		services.WriteResponseErr(&w, "can't decode request.", http.StatusBadRequest)
		return
	}
	if err := services.CreateShipping(ctx, req); err!=nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Failed create shipping. %v", err), http.StatusInternalServerError)
		return
	}
	services.WriteResponseSuccess(&w, "Success create shipping")
}

func GetShipping (w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	branch := mux.Vars(r)["branch"]
	limitStr := mux.Vars(r)["limit"]
	pageStr := mux.Vars(r)["page"]
	// Convert 'limit' and 'page' to integers
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		services.WriteResponseErr(&w, "Invalid or missing 'limit' parameter", http.StatusBadRequest)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		services.WriteResponseErr(&w, "Invalid or missing 'page' parameter", http.StatusBadRequest)
		return
	}

	shippings, err := services.GetShipping(ctx, branch, limit, page)
	if err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Failed get shipping. %v", err), http.StatusInternalServerError)
		return
	}
	services.WriteResponseSuccess(&w, shippings)
	
}