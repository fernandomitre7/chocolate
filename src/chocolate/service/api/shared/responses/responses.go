package responses

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"chocolate/service/api/shared/apierror"
)

const (
	mimeJSON = "application/json; charset=UTF-8"
	mimeHTML = "text/html; charset=UTF-8"
)

// Error returns an Error Response
func Error(r *http.Request, rw http.ResponseWriter, err *apierror.Error) {
	now := time.Now().UTC()
	// IMPORTANT NOTE: The client MUST NOT change its own time to the time returned by server
	//                 as it opens the possibility of some time attacks.
	rw.Header().Set("Date", fmt.Sprintf("%v", now.Format(http.TimeFormat)))
	rw.Header().Set("Server-Epoch", fmt.Sprintf("%v", now.Unix()))

	// Deactivating cache
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Expire", "Thu, 01 Dec 1994 16:00:00 GMT")
	rw.Header().Set("Pragma", "no-cache")
	// TODO Use Location
	respondJSON(rw, err.HTTPStatus, "/", err)
}

// Created returns 201 Created Resource
func Created(r *http.Request, rw http.ResponseWriter, v interface{}, location string) {
	now := time.Now().UTC()
	// IMPORTANT NOTE: The client MUST NOT change its own time to the time returned by server
	//                 as it opens the possibility of some time attacks.
	rw.Header().Set("Date", fmt.Sprintf("%v", now.Format(http.TimeFormat)))
	rw.Header().Set("Server-Epoch", fmt.Sprintf("%v", now.Unix()))

	// Deactivating cache
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Expire", "Thu, 01 Dec 1994 16:00:00 GMT")
	rw.Header().Set("Pragma", "no-cache")

	respondJSON(rw, http.StatusCreated, location, v)
}

// Ok returns 200 OK
func Ok(r *http.Request, rw http.ResponseWriter, v interface{}, location string) {
	now := time.Now().UTC()
	// IMPORTANT NOTE: The client MUST NOT change its own time to the time returned by server
	//                 as it opens the possibility of some time attacks.
	rw.Header().Set("Date", fmt.Sprintf("%v", now.Format(http.TimeFormat)))
	rw.Header().Set("Server-Epoch", fmt.Sprintf("%v", now.Unix()))

	// Deactivating cache
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Expire", "Thu, 01 Dec 1994 16:00:00 GMT")
	rw.Header().Set("Pragma", "no-cache")

	respondJSON(rw, http.StatusOK, location, v)
}

// NoContent returns 204 No Content
func NoContent(r *http.Request, rw http.ResponseWriter, location string) {
	now := time.Now().UTC()
	// IMPORTANT NOTE: The client MUST NOT change its own time to the time returned by server
	//                 as it opens the possibility of some time attacks.
	rw.Header().Set("Date", fmt.Sprintf("%v", now.Format(http.TimeFormat)))
	rw.Header().Set("Server-Epoch", fmt.Sprintf("%v", now.Unix()))

	// Deactivating cache
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Expire", "Thu, 01 Dec 1994 16:00:00 GMT")
	rw.Header().Set("Pragma", "no-cache")

	respondJSON(rw, http.StatusNoContent, location, nil)
}

// NotImplemented returns 501 Not Implemented
func NotImplemented(r *http.Request, rw http.ResponseWriter, location string) {
	now := time.Now().UTC()
	// IMPORTANT NOTE: The client MUST NOT change its own time to the time returned by server
	//                 as it opens the possibility of some time attacks.
	rw.Header().Set("Date", fmt.Sprintf("%v", now.Format(http.TimeFormat)))
	rw.Header().Set("Server-Epoch", fmt.Sprintf("%v", now.Unix()))

	// Deactivating cache
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Expire", "Thu, 01 Dec 1994 16:00:00 GMT")
	rw.Header().Set("Pragma", "no-cache")

	respondJSON(rw, http.StatusNotImplemented, location, nil)
}

// HTML returns an html page response
func HTML(r *http.Request, w http.ResponseWriter, reader io.Reader) {
	now := time.Now().UTC()
	// IMPORTANT NOTE: The client MUST NOT change its own time to the time returned by server
	//                 as it opens the possibility of some time attacks.
	w.Header().Set("Date", fmt.Sprintf("%v", now.Format(http.TimeFormat)))
	w.Header().Set("Server-Epoch", fmt.Sprintf("%v", now.Unix()))

	// Deactivating cache
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Expire", "Thu, 01 Dec 1994 16:00:00 GMT")
	w.Header().Set("Pragma", "no-cache")

	w.Header().Set("Content-Type", mimeHTML)
	w.WriteHeader(200)
	io.Copy(w, reader)
}

func respondJSON(w http.ResponseWriter, code int, location string, payload interface{}) {
	var apiresp APIResponse
	if apierr, ok := payload.(*apierror.Error); ok {
		apiresp = APIResponse{Error: apierr}
	} else {
		apiresp = APIResponse{Data: payload}
	}

	response, err := json.Marshal(apiresp)
	if err != nil {
		response = []byte(`{"error":"Error marshaling JSON response: ` + err.Error() + `"}`)
		code = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", mimeJSON)
	w.Header().Set("Location", location)
	//w.Write(response)
	w.WriteHeader(code)
	io.Copy(w, bytes.NewReader(response))
	//fmt.Fprintf(w, response)
}
