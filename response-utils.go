package devify

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ReadJSON reads a single JSON object from the request body into the provided data interface.
// It enforces a maximum body size of 1MB and ensures the body contains exactly one JSON value,
// rejecting requests with trailing data or multiple JSON objects.
//
// The data parameter must be a pointer to a struct where the JSON will be decoded.
// If strict is true, unknown fields in the JSON (relative to the data struct) are rejected;
// otherwise, they are ignored (default behavior). If the body is empty, it returns an error.
// If decoding fails (e.g., due to invalid JSON), a wrapped error is returned with details.
//
// Example:
//
//	type User struct {
//	    Name string `json:"name"`
//	}
//	var u User
//	err := d.ReadJSON(w, r, &u, true) // Strict mode: fails on unknown fields
//	if err != nil {
//	    // Handle error
//	}
//	err = d.ReadJSON(w, r, &u, false) // Lenient mode: ignores unknown fields
func (d *Devify) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}, strict bool) error {
	const maxBytes = 1048576 // 1MB limit
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	if strict {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(data); err != nil {
		if err == io.EOF {
			return errors.New("empty JSON body not allowed")
		}
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}
	return nil
}

// WriteJSON writes the provided payload as JSON to the http.ResponseWriter with the specified status code.
// It sets the Content-Type to "application/json" and applies any provided headers before writing the response.
// The JSON output is compact (no indentation) for efficiency.
//
// The payload can be any type that json.Marshal can handle (e.g., structs, maps, slices).
// Headers are optional and variadic; all provided header maps are merged into the response headers.
// Errors are returned if marshaling or writing fails.
//
// Example:
//
//	err := d.WriteJSON(w, http.StatusOK, map[string]string{"name": "Alice"})
//	if err != nil {
//	    // Handle error
//	}
//	// With headers:
//	h := http.Header{}
//	h.Set("X-Custom", "value")
//	err = d.WriteJSON(w, http.StatusCreated, struct{ ID int }{1}, h)
func (d *Devify) WriteJSON(w http.ResponseWriter, status int, payload interface{}, headers ...http.Header) error {
	out, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Apply all provided headers
	for _, header := range headers {
		for key, values := range header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
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

// WriteXML writes the provided payload as XML to the http.ResponseWriter with the specified status code.
// It sets the Content-Type to "application/xml" and applies any provided headers before writing the response.
// The XML output is compact (no indentation) for efficiency.
//
// The payload can be any type that xml.Marshal can handle (e.g., structs with xml tags, slices).
// Headers are optional and variadic; all provided header maps are merged into the response headers.
// Errors are returned if marshaling or writing fails.
//
// Example:
//
//	type User struct {
//	    XMLName xml.Name `xml:"user"`
//	    Name    string   `xml:"name"`
//	}
//	err := d.WriteXML(w, http.StatusOK, User{Name: "Alice"})
//	if err != nil {
//	    // Handle error
//	}
//	// With headers:
//	h := http.Header{}
//	h.Set("X-Custom", "value")
//	err = d.WriteXML(w, http.StatusCreated, User{Name: "Alice"}, h)
func (d *Devify) WriteXML(w http.ResponseWriter, status int, payload interface{}, headers ...http.Header) error {
	out, err := xml.Marshal(payload)
	if err != nil {
		return err
	}

	// Apply all provided headers
	for _, header := range headers {
		for key, values := range header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
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

// DownloadFile serves a file from the server as a downloadable attachment.
// It constructs the file path from pathToFile and fileName, ensures the path is safe,
// and sets the Content-Disposition header to trigger a download with the specified fileName.
//
// The pathToFile should be the base directory containing the file, and fileName is the name
// of the file within that directory. The function prevents path traversal by ensuring the
// resolved path remains within pathToFile. If the file cannot be accessed (e.g., not found
// or permission denied), an error is returned.
//
// Example:
//
//	err := d.DownloadFile(w, r, "/var/www/files", "report.pdf")
//	if err != nil {
//	    // Handle error, e.g., return 404
//	}
func (d *Devify) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) error {
	// Construct and clean the full file path
	fp := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(fp)

	// Security: Ensure the file is within the base directory
	if !strings.HasPrefix(fileToServe, filepath.Clean(pathToFile)) {
		return fmt.Errorf("invalid file path: %s attempts to access outside of %s", fileToServe, pathToFile)
	}

	// Check if the file exists and is readable
	if _, err := os.Stat(fileToServe); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", fileToServe)
	} else if err != nil {
		return fmt.Errorf("cannot access file %s: %w", fileToServe, err)
	}

	// Set the Content-Disposition header for download
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))

	// Serve the file
	http.ServeFile(w, r, fileToServe)
	return nil
}

func (d *Devify) Error404(w http.ResponseWriter) {
	d.ErrorSatus(w, http.StatusNotFound)
}

func (d *Devify) Error500(w http.ResponseWriter) {
	d.ErrorSatus(w, http.StatusInternalServerError)
}

func (d *Devify) ErrorUnauthorized(w http.ResponseWriter) {
	d.ErrorSatus(w, http.StatusUnauthorized)
}

func (d *Devify) ErrorForbidden(w http.ResponseWriter) {
	d.ErrorSatus(w, http.StatusForbidden)
}

func (d *Devify) ErrorBadRequest(w http.ResponseWriter) {
	d.ErrorSatus(w, http.StatusBadRequest)
}

func (d *Devify) ErrorTooManyRequests(w http.ResponseWriter) {
	d.ErrorSatus(w, http.StatusTooManyRequests)
}

func (d *Devify) ErrorPaymentRequired(w http.ResponseWriter) {
	d.ErrorSatus(w, http.StatusPaymentRequired)
}

func (d *Devify) ErrorSatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
	w.WriteHeader(status)
}
