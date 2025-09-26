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

func SendJSONResponse(w http.ResponseWriter, model interface{}) {
	json.NewEncoder(w).Encode(model)
}
