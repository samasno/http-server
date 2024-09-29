package httpserver

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ResponseWriter struct {
	version    string
	statusCode int
	message    string
	headers    http.Header
	writer     io.ReadWriter
}

func NewResponseWriter() *ResponseWriter {
	writer := bytes.NewBuffer([]byte{})
	return &ResponseWriter{
		writer:  writer,
		headers: http.Header{},
	}
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *ResponseWriter) Version(version string) {
	w.version = version
}

func (w *ResponseWriter) Header() http.Header {
	return w.headers
}

func (w *ResponseWriter) Marshal() []byte {
	buf := bytes.NewBuffer([]byte{})
	clrf := "\r\n"
	sp := " "
	buf.WriteString(w.version)
	buf.WriteString(sp)
	buf.WriteString(strconv.Itoa(w.statusCode))
	buf.WriteString(w.message)
	buf.WriteString(clrf)

	headers := w.Header()
	for k, v := range headers {
		buf.WriteString(k)
		buf.WriteString(":")
		buf.WriteString(strings.Join(v, ","))
		buf.WriteString(clrf)
	}
	buf.WriteString(clrf)
	io.Copy(buf, w.writer)

	return buf.Bytes()
}
