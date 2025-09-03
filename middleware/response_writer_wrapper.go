package middleware

import (
	"bytes"
	"net/http"
)

type ResponseWriterWrapper struct {
	w    *http.ResponseWriter
	body *bytes.Buffer
}

func NewResponseWriterWrapper(w http.ResponseWriter) ResponseWriterWrapper {
	var buf bytes.Buffer
	return ResponseWriterWrapper{
		w:    &w,
		body: &buf,
	}
}

func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
	rww.body.Write(buf)
	return (*rww.w).Write(buf)
}

func (rww ResponseWriterWrapper) Header() http.Header {
	return (*rww.w).Header()

}

func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
	(*rww.w).WriteHeader(statusCode)
}

func (rww ResponseWriterWrapper) GetBody() []byte {
	return rww.body.Bytes()
}
