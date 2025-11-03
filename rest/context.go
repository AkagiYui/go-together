// Package rest 提供轻量级的 RESTful API 框架
package rest

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"sync"
)

// Request HTTP 请求信息
type Request struct {
	Method   string // GET, POST, PUT, DELETE, ...
	Endpoint string // 不包含 query string
	URI      string // 包含 query string

	URL           url.URL
	Host          string
	RemoteAddr    string
	ContentLength int64

	// params in header
	Header     http.Header
	PathParams map[string]string
	Query      url.Values

	// body
	BodyType BodyType
	Form     url.Values
	Body     []byte
}

// Response HTTP 响应信息
type Response struct {
	StatusCode int
	Status     any
	Result     any
	Headers    http.Header
}

// Context HTTP 请求上下文
type Context struct {
	Request
	Response

	OriginalWriter  *http.ResponseWriter
	OriginalRequest *http.Request

	memoryLock sync.RWMutex
	Memory     map[any]any

	Server *Server

	currentRunnerIndex int           // 私有索引：当前执行位置
	runnerChain        []HandlerFunc // 当前请求的执行链

	disableInternalResponse bool
}

// SetStatus 设置响应状态
func (c *Context) SetStatus(status any) {
	c.Status = status
}

// SetStatusCode 设置 HTTP 状态码
func (c *Response) SetStatusCode(code int) {
	c.StatusCode = code
}

// SetResult 设置响应结果
func (c *Response) SetResult(result any) {
	c.Result = result
}

// Header 添加响应头
func (c *Response) Header(key, value string) {
	c.Headers.Add(key, value)
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key any) (value any, exists bool) {
	c.memoryLock.RLock()
	defer c.memoryLock.RUnlock()
	value, exists = c.Memory[key]
	return
}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key any, value any) {
	c.memoryLock.Lock()
	defer c.memoryLock.Unlock()
	c.Memory[key] = value
}

// Abort 中止后续处理器的执行
func (c *Context) Abort() {
	// Move index to the end to ensure subsequent Next does not execute remaining handlers
	c.currentRunnerIndex = len(c.runnerChain)
}

// IsAborted 检查是否已中止
func (c *Context) IsAborted() bool {
	return c.currentRunnerIndex >= len(c.runnerChain)
}

// Next executes the remaining handlers in the chain starting from the current index
func (c *Context) Next() {
	for c.currentRunnerIndex++; c.currentRunnerIndex < len(c.runnerChain); c.currentRunnerIndex++ {
		c.runnerChain[c.currentRunnerIndex](c)
	}
}

// NewContext 创建一个新的请求上下文
func NewContext(r *http.Request, w *http.ResponseWriter, s *Server, runnerChain []HandlerFunc) *Context {
	ctx := &Context{
		Request: Request{
			Method: r.Method,

			Endpoint: r.URL.Path,
			URI:      r.RequestURI,

			URL:           *r.URL,
			Host:          r.Host,
			RemoteAddr:    r.RemoteAddr,
			ContentLength: r.ContentLength,

			Header:     r.Header,
			PathParams: make(map[string]string),
			Query:      r.URL.Query(),

			BodyType: Nil,
			Form:     nil, // Form 暂不处理
			Body:     nil,
		},
		Response: Response{
			Status:     nil,
			StatusCode: http.StatusOK,
			Result:     nil,
			Headers:    make(http.Header),
		},

		OriginalWriter:  w,
		OriginalRequest: r,

		memoryLock: sync.RWMutex{},
		Memory:     make(map[any]any),

		Server: s,

		currentRunnerIndex: -1,
		runnerChain:        runnerChain,

		disableInternalResponse: false,
	}

	// 解析请求体类型
	contentType, _, err := mime.ParseMediaType(ctx.Request.Header.Get("Content-Type"))

	if err != nil {
		if !slices.Contains([]string{http.MethodGet}, ctx.Method) && !strings.Contains(err.Error(), "no media type") {
			fmt.Printf("Method: %s, Content-Type: %s\n", ctx.Method, contentType)
			ctx.SetStatusCode(http.StatusBadRequest)
			ctx.SetResult("Invalid Content-Type")
			ctx.Abort()
			return ctx
		}
	}
	switch contentType {
	case "application/x-www-form-urlencoded":
		ctx.BodyType = EncodeURL
	case "application/json":
		ctx.BodyType = JSON
	case "multipart/form-data":
		ctx.BodyType = FormData
	}

	// 解析路径参数
	keys, values := parsePathParams(r)
	for i, key := range keys {
		ctx.PathParams[key] = values[i]
	}

	return ctx
}

func parsePathParams(r *http.Request) (keys []string, values []string) {
	// 反射获取 Request 中的 pat(*net/http.pattern) 和 matches([]string)
	keys = make([]string, 0)
	values = make([]string, 0)

	t := reflect.TypeOf(r).Elem() // http.Request
	v := reflect.ValueOf(r).Elem()

	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name

		// 获取 路径参数 值
		if fieldName == "matches" {
			fieldValue := v.Field(i) // 获取 matches ([]string)
			if fieldValue.IsValid() && fieldValue.Kind() == reflect.Slice {
				for j := 0; j < fieldValue.Len(); j++ {
					elem := fieldValue.Index(j)
					if elem.Kind() == reflect.String {
						values = append(values, elem.String())
					}
				}
			}
			continue
		}

		// 获取 路径参数 键 pat(*net/http.pattern){segments([]net/http.segment):[{s string, wild bool}}]
		if fieldName == "pat" {
			fieldValue := v.Field(i) // 获取 pat (*net/http.pattern)
			if fieldValue.IsValid() && fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
				patternValue := fieldValue.Elem() // 获取 net/http.pattern
				patternType := patternValue.Type()

				// 查找 segments 字段
				for j := 0; j < patternType.NumField(); j++ {
					if patternType.Field(j).Name == "segments" {
						segmentsValue := patternValue.Field(j)
						if segmentsValue.IsValid() && segmentsValue.Kind() == reflect.Slice {
							// 遍历每个 segment
							for k := 0; k < segmentsValue.Len(); k++ {
								segmentValue := segmentsValue.Index(k)
								segmentType := segmentValue.Type()

								var s string
								var wild bool

								// 获取 segment 的 s 和 wild 字段
								for l := 0; l < segmentType.NumField(); l++ {
									fieldName := segmentType.Field(l).Name
									fieldVal := segmentValue.Field(l)

									if fieldName == "s" && fieldVal.Kind() == reflect.String {
										s = fieldVal.String()
									} else if fieldName == "wild" && fieldVal.Kind() == reflect.Bool {
										wild = fieldVal.Bool()
									}
								}

								// 如果是通配符段，添加到 pathKeys
								if wild && s != "" {
									keys = append(keys, s)
								}
							}
						}
						break
					}
				}
			}
			continue
		}
	}
	return
}

// FillBody 读取请求体并缓存到 ctx.Body
func (c *Context) FillBody() []byte {
	if c.Body == nil {
		body, err := io.ReadAll(c.OriginalRequest.Body)
		if err != nil {
			panic(err)
		}
		c.Body = body
		// 重置原请求体
		c.OriginalRequest.Body = io.NopCloser(io.MultiReader(c.OriginalRequest.Body, io.NopCloser(bytes.NewReader(c.Body))))
	}
	return c.Body
}

// Stream 流式响应
func (c *Context) Stream(step func(w io.Writer) bool) bool {
	c.disableInternalResponse = true
	c.Response.Headers.Add("Transfer-Encoding", "chunked")
	c.writeHeaders()

	w := *c.OriginalWriter
	clientGone := c.OriginalRequest.Context().Done()
	for {
		select {
		case <-clientGone:
			return true
		default:
			keepOpen := step(w)
			w.(http.Flusher).Flush()
			if !keepOpen {
				return false
			}
		}
	}
}

func (c *Context) writeHeaders() {
	for key, values := range c.Response.Headers {
		for _, value := range values {
			(*c.OriginalWriter).Header().Add(key, value)
		}
	}
}

// DisableInternalResponse 禁用内部响应处理
func (c *Context) DisableInternalResponse() {
	c.disableInternalResponse = true
}

// NewEmptyContext 创建一个空的上下文实例
func NewEmptyContext() Context {
	return Context{
		disableInternalResponse: true,
	}
}
