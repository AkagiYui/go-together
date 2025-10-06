package rest

import (
	"net/http"
	"net/url"
)

type Context struct {
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

	header     http.Header
	statusCode int
}

func (c *Context) Status(code int) {
	c.statusCode = code
}
