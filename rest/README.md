一个对 Go1.22 ServeMux 进行封装，我自己用的 [REST](https://en.wikipedia.org/wiki/REST) 框架。

支持路径参数、查询参数、请求头、JSON请求体、表单参数等的自动绑定。

路由语法与 ServeMux 保持一致，本库未重定义路由语法。

```golang
package main

import "github.com/akagiyui/go-together/rest"

type HelloRequest struct {
	AccessToken string `header:"access-token"`
	Type        string `path:"type"`
	PathSuffix  string `path:"suffix"`
	Name        string `query:"name"`
	Age         int    `json:"age"`
}

func (r *HelloRequest) Handle(ctx *rest.Context) {
	fmt.Printf("req: %v\n", r)
	ctx.Result("zgm\n")
}

func main() {
	s := rest.NewServer()

	s.GETFunc("/healthz", func(ctx *rest.Context) {
		ctx.Result("OK")
	})
    s.POST("/hello/{type}/zgm/{suffix...}", &HelloRequest{}) // go1.22

	if err := s.Run(":8080"); err != nil {
		panic(err)
	}
}
```

```shell
curl -X POST "http://localhost:8080/hello/123/zgm/456/ddi.txt?name=akagi" \
-H "Access-Token: 123" \
-H "Content-Type: application/json" \
-d '{"age": 18}'
```
