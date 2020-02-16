package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"raffle/knowledge/pkg"
	"runtime"
	"strconv"
	"strings"
)

// handler represents an implementation of the handler interface
type handler struct {
	documentService pkg.DocumentService
	logger          *log.Logger
}

// NewHandler returns a poitner to a handler
func NewHandler(documentService pkg.DocumentService, logger *log.Logger) *handler {
	h := &handler{
		documentService: documentService,
		logger:          logger,
	}
	return h
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/documents") {
		switch r.Method {
		case "GET":
			h.getAll(w, r)
		}
	} else if strings.HasPrefix(r.URL.Path, "/document") {
		switch r.Method {
		case "GET":
			h.getOne(w, r)
		case "POST":
			h.postOne(w, r)
		case "PUT":
			h.putOne(w, r)
		case "DELETE":
			h.deleteOne(w, r)
		}
	} else {
		http.NotFound(w, r)
	}
}

func (h handler) getOne(w http.ResponseWriter, r *http.Request) {
	// Retrieve doc_id parameter and convert to int.
	// Error if left out of int.
	docIDStr := r.FormValue("doc_id")
	if docIDStr == "" {
		h.error(w, errors.New("missing doc_id parameter"), http.StatusBadRequest)
		return
	}
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		h.error(w, errors.New("doc_id parameter must be of type int"), http.StatusBadRequest)
		return
	}
	// Retrieve doc_id parameter if there and convert
	// to int. Else set to 0. Error if not int.
	docVersionStr := r.FormValue("doc_version")
	var docVersion int
	if docVersionStr == "" {
		docVersion = 0
	} else {
		docVersion, err = strconv.Atoi(docVersionStr)
		if err != nil {
			h.error(w, errors.New("doc_version parameter must be of type int or left out"), http.StatusBadRequest)
			return
		}
	}
	// Retrieve document.
	doc, err := h.documentService.RetrieveDocument(docID, docVersion)
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
	} else if doc == nil {
		h.notFound(w)
	} else {
		h.encodeJSON(w, doc)
	}
}

func (h handler) postOne(w http.ResponseWriter, r *http.Request) {
	// Decode ClientDocument from request.
	// Error if invalid json document.
	doc := &pkg.ClientDocument{}
	err := json.NewDecoder(r.Body).Decode(doc)
	if err != nil {
		h.error(w, errors.New("invalid json"), http.StatusBadRequest)
		return
	}
	// Store document.
	version, err := h.documentService.StoreDocument(doc)
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
	} else {
		h.encodeJSON(w, postResponse{version})
	}
}

func (h handler) putOne(w http.ResponseWriter, r *http.Request) {
	// Retrieve doc_version parameter and convert to int.
	// Error if left out of int.
	docVersionStr := r.FormValue("doc_version")
	if docVersionStr == "" {
		h.error(w, errors.New("missing doc_version parameter"), http.StatusBadRequest)
		return
	}
	docVersion, err := strconv.Atoi(docVersionStr)
	if err != nil {
		h.error(w, errors.New("doc_version parameter must be of type int"), http.StatusBadRequest)
		return
	}
	// Decode ClientDocument from request.
	// Error if invalid json document.
	doc := &pkg.ClientDocument{}
	err = json.NewDecoder(r.Body).Decode(doc)
	if err != nil {
		h.error(w, errors.New("invalid json"), http.StatusBadRequest)
		return
	}
	// Update document.
	err = h.documentService.UpdateDocument(doc, docVersion)
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
		return
	}
}

func (h handler) deleteOne(w http.ResponseWriter, r *http.Request) {
	// Retrieve doc_id parameter and convert to int.
	// Error if left out of string type.
	docIDStr := r.FormValue("doc_id")
	if docIDStr == "" {
		h.error(w, errors.New("missing doc_id parameter"), http.StatusBadRequest)
		return
	}
	docID, err := strconv.Atoi(docIDStr)
	if err != nil {
		h.error(w, errors.New("doc_id parameter must be of type int"), http.StatusBadRequest)
		return
	}
	// Retrieve doc_id parameter if there and convert
	// to int. Else set to 0. Error if not int.
	docVersionStr := r.FormValue("doc_version")
	var docVersion int
	if docVersionStr == "" {
		docVersion = 0
	} else {
		docVersion, err = strconv.Atoi(docVersionStr)
		if err != nil {
			h.error(w, errors.New("doc_version parameter must be of type int or left out"), http.StatusBadRequest)
			return
		}
	}
	// Delete document.
	err = h.documentService.RemoveDocument(docID, docVersion)
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
	}
}

func (h handler) getAll(w http.ResponseWriter, r *http.Request) {
	// Retrieve doc_version parameter and convert to int.
	// Error if left out of string type.
	docVersionStr := r.FormValue("doc_version")
	var docVersion int
	var err error
	if docVersionStr == "" {
		h.error(w, errors.New("missing doc_version parameter"), http.StatusBadRequest)
		return
	}
	docVersion, err = strconv.Atoi(docVersionStr)
	if err != nil {
		h.error(w, errors.New("doc_version parameter must be of type int"), http.StatusBadRequest)
		return
	}
	// Retrieve all documents.
	docs, err := h.documentService.RetrieveDocuments(docVersion)
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
	} else if docs == nil {
		h.notFound(w)
	} else {
		h.encodeJSON(w, docs)
	}
}

// Error writes an API error message to the response and Logger.
func (h handler) error(w http.ResponseWriter, err error, code int) {
	// Log error.
	_, fn, line, _ := runtime.Caller(1)
	h.logger.Printf("http error: %s (code=%d,file=%s, line=%d)", err, code, fn, line)
	// Hide error from client if it is internal.
	if code == http.StatusInternalServerError {
		err = errors.New("internal server error. please contact the api provider")
	}
	// Write generic error response.
	w.WriteHeader(code)
	err = json.NewEncoder(w).Encode(errResponse{Err: err.Error()})
	if err != nil {
		h.error(w, err, http.StatusInternalServerError)
	}
}

// NotFound writes an API error message to the response.
func (h handler) notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{}` + "\n"))
}

// encodeJSON encodes v to w in JSON format. Error() is called if encoding fails.
func (h handler) encodeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.error(w, err, http.StatusInternalServerError)
	}
}

type postResponse struct {
	Version int `json:"version"`
}

type errResponse struct {
	Err string `json:"error"`
}
