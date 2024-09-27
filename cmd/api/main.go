package main

import (
	"TouchySarun/chp_order_backend/internal/firestore"
	"TouchySarun/chp_order_backend/internal/handlers"
	"TouchySarun/chp_order_backend/internal/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Function to define routes
func registerRoutes(router *mux.Router) {
	router.HandleFunc("/api/users/{id}", handlers.GetUserById).Methods("GET") // 
	router.HandleFunc("/api/users", handlers.GetUsers).Methods("GET") // Get all users
	router.HandleFunc("/api/users", handlers.CreateUser).Methods("POST") // Create user {body: {username, password}}
	router.HandleFunc("/api/sku/{barcode}", handlers.GetSku).Methods("GET")
	router.HandleFunc("/api/orders/create-data/{barcode}/{branch}", handlers.GetCreateOrderData).Methods("GET") // Get create order data
	router.HandleFunc("/api/orders", handlers.CreateOrder).Methods("POST") // Create order {body:{branch, name,utqName,utqQty,code,sku,ap,qty,cat,bnd,creBy}}
	router.HandleFunc("/api/orders", handlers.EditOrder).Methods("PUT") // Edit order {id, qty, utqName, utqqty, code, creBy}
	router.HandleFunc("/api/orders", handlers.GetOrders).Methods("GET") // Get orders with query params
	router.HandleFunc("/api/orders/{id}", handlers.UpdateStatus).Methods("POST") // Update order status {creBy, qty, status}
	router.HandleFunc("/api/shipping", handlers.CreateShipping).Methods("POST") // Create temp shipping {orderId, qty, branch} // qty = dif then all is edit
	router.HandleFunc("/api/shipping/confirm", handlers.ConfirmShipping).Methods("POST") // Confirm shipping {branch, creBy}
	router.HandleFunc("/api/shipping/{branch}/{limit}/{page}", handlers.GetShipping).Methods("GET") // Get temp shipping // qty = sum of all orderId
}

//TODO: limit, offset = params or query

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found")
	}

	// Set up firestore
	firestore.InitFirestore()
	defer firestore.Client.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set
	}

	log.Printf("Starting server on port %s ...", port)

	router := mux.NewRouter()

	// Register routes
	registerRoutes(router)

	// Add middleware
	router.Use(middleware.LoggingMiddleware) // Log every request
	router.Use(middleware.RecoverMiddleware) // Handle panics and recover gracefully

	// Start server
	log.Fatal(http.ListenAndServe(":"+port, router))
}