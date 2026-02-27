package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// JSONRequest creates an HTTP request with JSON body
func JSONRequest(method, path string, body interface{}) *http.Request {
	var bodyReader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Recorder wraps httptest.ResponseRecorder for convenience
func Recorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}
