package rest

import (
	"net/http"
	"net/url"
)

type Request struct {
	Method string

	Endpoint string // 不包含 query string
	URI      string

	Body          []byte
	URL           url.URL
	Host          string
	RemoteAddr    string
	ContentLength int64

	Params map[string]string
	Query  map[string]string
	Header map[string]string
	Form   map[string]string

	header http.Header
}

type Response struct {
	statusCode int
}

type Context struct {
	Request
	Response
}

func (c *Context) Status(code int) {
	c.statusCode = code
}

func NewContext(r *http.Request) *Context {
	return &Context{
		Request: Request{
			Method: r.Method,

			Endpoint: r.URL.Path,
			URI:      r.RequestURI,

			Body:          nil,
			URL:           *r.URL,
			Host:          r.Host,
			RemoteAddr:    r.RemoteAddr,
			ContentLength: r.ContentLength,

			Params: make(map[string]string),
			Header: make(map[string]string),
			Query:  make(map[string]string),
			Form:   make(map[string]string),

			header: r.Header,
		},
		Response: Response{
			statusCode: http.StatusOK,
		},
	}
}
