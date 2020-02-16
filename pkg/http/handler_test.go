package http

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"raffle/knowledge/pkg/mongo"
	"strings"
	"testing"
)

var H *handler

func init() {
	H = NewTestHandler()
}

func TestHandler(t *testing.T) {
	t.Run("POST document", testHandler_postOne)
	t.Run("POST document with error bad request", testHandler_postOne_ErrBadRequest)
	t.Run("GET document", testHandler_getOne)
	t.Run("GET document with error bad request", testHandler_getOne_ErrBadRequest)
	t.Run("GET document with error not found", testHandler_getOne_ErrNotFound)
	t.Run("PUT document", testHandler_putOne)
	t.Run("DELETE document", testHandler_deleteOne)
}

func testHandler_getOne(t *testing.T) {
	r, err := http.NewRequest("GET", "localhost:8080/document?doc_id=1", nil)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	H.getOne(w, r)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
}

func testHandler_getOne_ErrNotFound(t *testing.T) {
	r, err := http.NewRequest("GET", "localhost:8080/document?doc_id=0", nil)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	H.getOne(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}
}

func testHandler_getOne_ErrBadRequest(t *testing.T) {
	r, err := http.NewRequest("GET", "localhost:8080/document?doc_id=1&doc_version=abc", nil)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	H.getOne(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
}

func testHandler_postOne(t *testing.T) {
	b := []byte(`{"doc_id": 1,"doc_url": "www.test.com","elements":[{"text":"testing 123","type":"h2"}]}`)
	r, err := http.NewRequest("POST", "localhost:8080/document", bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	H.postOne(w, r)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
}

func testHandler_postOne_ErrBadRequest(t *testing.T) {
	b := []byte(`{"invalid_field","doc_id": 1,"doc_url": "www.test.com","elements":[{"text":"testing 123","type":"h2"}]}`)
	r, err := http.NewRequest("POST", "localhost:8080/document", bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	H.postOne(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatal(w.Code)
	}
}

func testHandler_putOne(t *testing.T) {
	v := "1"
	bIn := []byte(`{"doc_id":1,"doc_url":"www.test.com","elements":[{"text":"testing 123","type":"h4"}]}`)
	r, err := http.NewRequest("PUT", "localhost:8080/document?doc_version="+v, bytes.NewReader(bIn))
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	H.putOne(w, r)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	r, err = http.NewRequest("GET", "localhost:8080/document?doc_id=1&doc_version="+v, nil)
	if err != nil {
		log.Fatal(err)
	}
	w = httptest.NewRecorder()
	H.getOne(w, r)
	bOut, err := ioutil.ReadAll(w.Body)
	if err != nil {
		log.Fatal(err)
	}
	if strings.TrimRight(string(bIn), "\n") != strings.TrimRight(string(bOut), "\n") {
		t.Fatal(string(bOut) + " != " + string(bIn))
	}
}

func testHandler_deleteOne(t *testing.T) {
	docID := "1"
	r, err := http.NewRequest("DELETE", "localhost:8080/document?doc_id="+docID, nil)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	H.deleteOne(w, r)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	r, err = http.NewRequest("GET", "localhost:8080/document?doc_id="+docID, nil)
	if err != nil {
		log.Fatal(err)
	}
	w = httptest.NewRecorder()
	H.getOne(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}
}

func NewTestHandler() *handler {
	f, err := os.OpenFile("test.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	l := log.New(f, "", log.Ldate|log.Ltime|log.Lshortfile)
	ds, err := mongo.NewDocumentService("mongodb://localhost:27017", "knowledge", "one", l)
	if err != nil {
		log.Fatal(err)
	}
	h := NewHandler(ds, l)
	return h
}
