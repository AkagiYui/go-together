package rest

type Context struct {
	Method string
	Path   string
	Body   []byte

	statusCode int
}

func (c *Context) Status(code int) {
	c.statusCode = code
}
