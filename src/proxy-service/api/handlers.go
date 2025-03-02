package api

import (
	"net/http"
)

func ProxySpecificHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"This is handled by the proxy service directly"}`))
}
