package listener

import (
	"net/http"
)

var r Record

// Writer is a middleware handler that writes to db
type Writer struct {
	handler http.Handler
}

// NewWriter constructs a new middleware handler
func NewWriter(handlerToWrap http.Handler) *Writer {
	return &Writer{handlerToWrap}
}

// serveHTTP passes the request from main handler to middleware
func (wr *Writer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	wr.handler.ServeHTTP(rw, req)

	err := recorder(&r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

// Handler serves a wrapped mux
func Handler() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.ParseHandler)
	return NewWriter(mux)
}
