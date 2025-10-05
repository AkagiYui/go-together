package rest

type Context struct {
	Method string
	Path   string
	Body   []byte
}
