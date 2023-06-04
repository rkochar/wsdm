package shared

import (
	"log"
	"net/http"
	"strconv"
)

const API_GATEWAY = "http://api-gateway-service-0:5000/orders/checkout/"

func RouteCheckoutCall(orderID string, status int) int {
	str := strconv.Itoa(status)
	backendURL := API_GATEWAY + orderID + "/" + str
	log.Printf("Calling %s", backendURL)
	resp, err := http.Get(backendURL)
	if err != nil {
		log.Printf("\nFailed to make service call: %v", err)
		return http.StatusBadRequest
	}
	log.Print(resp.StatusCode)
	return resp.StatusCode
}
