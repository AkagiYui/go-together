package route

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterRainApiProxyRoutes(r *gin.Engine) {
	target := "https://api.v2.rainyun.com"
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("解析目标URL失败: %v\n", err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.Host = "api.v2.rainyun.com"
		req.URL.Scheme = "https"
		req.URL.Host = "api.v2.rainyun.com"
		// 路径删除 /proxy/rain/api 前缀
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/proxy/rain/api")

		// 删除 cookie
		req.Header.Del("Cookie")
		// 设置referer
		req.Header.Set("Referer", "https://app.rainyun.com/")

		// 删除 x-forwarded开头的头部
		for k := range req.Header {
			if strings.HasPrefix(k, "X-Forwarded") {
				req.Header.Del(k)
			}
		}
	}
	// 添加 ModifyResponse 来移除上游服务器的 CORS 头
	proxy.ModifyResponse = func(r *http.Response) error {
		// 删除上游服务器的 CORS 相关头部
		r.Header.Del("Access-Control-Allow-Origin")
		r.Header.Del("Access-Control-Allow-Credentials")
		r.Header.Del("Access-Control-Allow-Methods")
		r.Header.Del("Access-Control-Allow-Headers")

		// 移除 set-cookie 头
		r.Header.Del("Set-Cookie")

		// 打印代理实际发送的请求url和请求头
		log.Printf("Proxy request URL: %s\n", r.Request.URL.String())
		log.Printf("Proxy request headers: %v\n", r.Request.Header)

		return nil
	}

	// 创建反向代理处理函数
	r.Any("/proxy/rain/api/*proxyPath", func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
