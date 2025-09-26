package utils

import (
	"encoding/json"
	"net/http"
)

func ReadJSONRequest(r *http.Request, model interface{}) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(model); err != nil {
		return err
	}
	return nil
}

func SendJSONResponse(w http.ResponseWriter, model interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(model)
}
