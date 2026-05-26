package api

import (
	"encoding/json"
	"net/http"
)

// вынесена функция JSONHandler в api.go для удобства(используется в redirect/, deleter/)

type Response struct {
	Error  string `json:"error,omitempty"`
	Status string `json:"status"`
}

func JSONHandler(w http.ResponseWriter, error string, status string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(Response{
		Error:  error,
		Status: status,
	})
	if err != nil {
		return
	}
}
