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
// if else "GET, POST" must have OPTION after
func registerRoutes(router *mux.Router) {
	router.HandleFunc("/api/login", handlers.Login).Methods("POST")
	router.HandleFunc("/api/users/{id}", handlers.GetUserById).Methods("GET") // 
	router.HandleFunc("/api/users/{id}", handlers.EditUser).Methods("PUT", http.MethodOptions) //
	router.HandleFunc("/api/users", handlers.CreateUser).Methods("POST") // Create user {body: {username, password}}
	router.HandleFunc("/api/sku/{barcode}", handlers.GetSku).Methods("GET")
	router.HandleFunc("/api/orders/create-data/{barcode}/{branch}", handlers.GetCreateOrderData).Methods("GET") // Get create order data
	router.HandleFunc("/api/orders", handlers.CreateOrder).Methods("POST") // Create order {body:{branch, name,utqName,utqQty,code,sku,ap,qty,cat,bnd,creBy}}
	router.HandleFunc("/api/orders", handlers.EditOrder).Methods("PUT", http.MethodOptions) // Edit order {id, qty, utqName, utqqty, code, creBy}
	router.HandleFunc("/api/orders", handlers.GetOrders).Methods("GET") // Get orders with query params
	router.HandleFunc("/api/orders/{id}", handlers.UpdateStatus).Methods("POST") // Update order status {creBy, qty, status}
	router.HandleFunc("/api/shipping", handlers.CreateShipping).Methods("POST") // Create temp shipping {orderId, qty, branch} // qty = dif then all is edit
	router.HandleFunc("/api/shipping/confirm", handlers.ConfirmShipping).Methods("POST") // Confirm shipping {branch, creBy}
	router.HandleFunc("/api/shipping/{branch}/{limit}/{page}", handlers.GetShipping).Methods("GET") // Get temp shipping // qty = sum of all orderId
	router.HandleFunc("/setup/skus", handlers.CreateSkus).Methods("POST") // Create sku from csv file
}

//TODO: limit, offset = params or query

// CORS middleware to handle CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins; adjust as needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r) // Call the next handler
	})
}


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
	// router.Use(mux.CORSMethodMiddleware(router))
	router.Use(corsMiddleware) 
	router.Use(middleware.LoggingMiddleware) // Log every request
	router.Use(middleware.RecoverMiddleware) // Handle panics and recover gracefully
	// Start server
	log.Fatal(http.ListenAndServe(":"+port, router))
}