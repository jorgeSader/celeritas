package celeritas

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
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

func (c *Celeritas) WriteXML(w http.ResponseWriter, status int, payload interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) error {
	fp := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	http.ServeFile(w, r, fileToServe)
	return nil
}

func (c *Celeritas) Error404(w http.ResponseWriter) {
	c.ErrorSatus(w, http.StatusNotFound)
}

func (c *Celeritas) Error500(w http.ResponseWriter) {
	c.ErrorSatus(w, http.StatusInternalServerError)
}

func (c *Celeritas) ErrorUnauthorized(w http.ResponseWriter) {
	c.ErrorSatus(w, http.StatusUnauthorized)
}

func (c *Celeritas) ErrorForbidden(w http.ResponseWriter) {
	c.ErrorSatus(w, http.StatusForbidden)
}

func (c *Celeritas) ErrorBadRequest(w http.ResponseWriter) {
	c.ErrorSatus(w, http.StatusBadRequest)
}

func (c *Celeritas) ErrorTooManyRequests(w http.ResponseWriter) {
	c.ErrorSatus(w, http.StatusTooManyRequests)
}

func (c *Celeritas) ErrorPaymentRequired(w http.ResponseWriter) {
	c.ErrorSatus(w, http.StatusPaymentRequired)
}

func (c *Celeritas) ErrorSatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
	w.WriteHeader(status)
}
