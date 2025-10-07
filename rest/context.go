package rest

import (
	"net/http"
	"net/url"
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

	PathValues url.Values
	Query      url.Values
	Header     http.Header

	Form url.Values
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

			PathValues: make(url.Values),
			Header:     r.Header,
			Query:      r.URL.Query(),

			Form: nil, // Form 暂不处理
		},
		Response: Response{
			statusCode: http.StatusOK,
		},
		OriginalWriter:  w,
		OriginalRequest: r,
	}

	return ctx
}
