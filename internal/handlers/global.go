package handlers

import (
	"TouchySarun/chp_order_backend/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
)

func GetBranches (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	branches, err := services.GetBranches(ctx)
	if err != nil {
		services.WriteResponseErr(&w, fmt.Sprintf("Failed, Getting branches, %v",err),http.StatusInternalServerError)
	}
	services.WriteResponseSuccess(&w,branches)
}

func CreateBranch (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type createBranchBody struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	var req createBranchBody 
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		services.WriteResponseErr(&w, "can not decode request.", http.StatusBadRequest)
	}
	

	if err := services.CreateBranch(ctx, req.Code, req.Name); err != nil {
		services.WriteResponseErr(&w, err.Error(), http.StatusInternalServerError)
	}
	services.WriteResponseSuccess(&w, "success create new branch")
}