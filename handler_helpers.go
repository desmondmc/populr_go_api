package main

import (
	"encoding/json"
	"net/http"
)

func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	if obj, ok := data.(Public); ok {
		data = obj.Public()
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&Resource{Data: data})
}

func OptionsRespond(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(200)
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}
