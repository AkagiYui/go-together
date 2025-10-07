package rest

import (
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	Method string // GET, POST, PUT, DELETE, ...

	Endpoint string // 不包含 query string
	URI      string // 包含 query string

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
	result     any
}

type Context struct {
	Request
	Response

	OriginalWriter  *http.ResponseWriter
	OriginalRequest *http.Request
}

func (c *Response) Status(code int) {
	c.statusCode = code
}

func (c *Response) Result(result any) {
	c.result = result
}

func NewContext(r *http.Request, w *http.ResponseWriter) *Context {
	ctx := &Context{
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
		OriginalWriter:  w,
		OriginalRequest: r,
	}

	// 填充 Header（多值以逗号拼接）
	for k, vs := range r.Header {
		if len(vs) > 0 {
			ctx.Request.Header[k] = strings.Join(vs, ",")
		}
	}

	// 填充 Query（多值以逗号拼接）
	q := r.URL.Query()
	for k, vs := range q {
		if len(vs) > 0 {
			ctx.Request.Query[k] = strings.Join(vs, ",")
		}
	}
	
	return ctx
}
