package httpserver

import (
	"io"
	"net/http"
)

func HandleRequest(c *Conn) (*http.Request, error) {
	data := []byte{}
	b := make([]byte, 5*1024)

	n, err := c.Read(b)
	if err != nil && err != io.EOF {
		return nil, err
	}

	data = append(data, b[:n]...)

	r, err := parseHttpRequest(data, c)
	if err != nil {
		return nil, err
	}
	return r, nil
}

type Handler struct {
	routes map[string]http.Handler
}

func NewHandler() *Handler {
	h := &Handler{
		routes: map[string]http.Handler{},
	}

	return h
}

func (h *Handler) Handle(pattern string, handler http.Handler) {
	h.routes[pattern] = handler
}

func (h *Handler) HandlerFunc(pattern string, handler http.HandlerFunc) {
	h.routes[pattern] = handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, ok := h.routes[r.URL.Path]
	println("found handler")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	handler.ServeHTTP(w, r)
}
