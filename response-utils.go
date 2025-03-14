package celeritas

import (
	"encoding/json"
	"net/http"
)

func (c *Celeritas) WriteJSON(w http.ResponseWriter, status int, payload interface{}, headers ...http.Header) error {
	//TODO: change to json.Marshal in production.
	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}
