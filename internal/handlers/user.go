package handlers

import (
	"encoding/json"
	"fmt"
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
		}
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["id"]
	ctx := r.Context()
	fmt.Println("GET user by id")
	user, err := services.GetUser(ctx, userId)
	if err != nil {
		services.WriteResponseErr(&w, "User not found.", http.StatusNotFound)
	} else {
		services.WriteResponseSuccess(&w, user)
	}
}

func GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	ctx := r.Context()

	user, err := services.GetUserByUsername(ctx, username)
	if err != nil {
		services.WriteResponseErr(&w, "User not found.", http.StatusNotFound)
	} else {
		services.WriteResponseSuccess(&w, user)
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

func Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	type Login struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var req Login
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	user, err := services.GetUserByUsername(ctx, req.Username)
	if err != nil {
		services.WriteResponseErr(&w, "User not found.", http.StatusNotFound)
	} else if user.Password != req.Password{
		services.WriteResponseErr(&w, "Incorrect password.", http.StatusNotFound)
	} else {
		services.WriteResponseSuccess(&w, user)
	}
}