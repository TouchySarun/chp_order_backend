package handlers

import (
	"encoding/json"
	"net/http"

	"TouchySarun/chp_order_backend/internal/models"
	"TouchySarun/chp_order_backend/internal/services"

	"github.com/gorilla/mux"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	// 	 users, err := services.GetAllUsers()
  //   if err != nil {
  //       http.Error(w, "Error fetching users", http.StatusInternalServerError)
  //       return
  //   }
    users := models.User{
			Name: "John doe",
			Age: 45,
		}
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]

	ctx := r.Context()
	userData, err := services.GetUser(ctx, userId)
	if err != nil {
		http.Error(w, "User not found.", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // err := services.CreateUser(user)
    // if err != nil {
    //     http.Error(w, "Error creating user", http.StatusInternalServerError)
    //     return
    // }

    w.WriteHeader(http.StatusCreated)
}