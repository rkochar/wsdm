package main

//
//import (
//	"fmt"
//	"github.com/gorilla/mux"
//	"log"
//	"net/http"
//)
//
//func main() {
//	// Define the server port
//	port := "5002"
//
//	// Set up a route handler function
//	router := mux.NewRouter()
//	router.HandleFunc("/orders/checkout/{order_id}", greeting)
//
//	fmt.Printf("Current port is: %s\n", port)
//	if port == "" {
//		port = "5002"
//	}
//
//	// Set the listening address and port for the server
//	addr := fmt.Sprintf(":%s", port)
//	fmt.Printf("Starting order service at %s\n", addr)
//	log.Fatal(http.ListenAndServe(addr, router))
//}
//
//func greeting(w http.ResponseWriter, r *http.Request) {
//	// Process the request and generate the response
//
//	// Set the response status code
//	w.WriteHeader(http.StatusOK)
//
//	// Write the response body
//	fmt.Fprint(w, "Checkout successful!")
//}
