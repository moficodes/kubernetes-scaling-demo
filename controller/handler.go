package main

import (
	"encoding/json"
	"net/http"
)

func getInstances(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		e := json.NewEncoder(w)
		e.Encode(instances)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
