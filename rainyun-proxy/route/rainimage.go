package route

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterRainImageProxyRoutes 注册用于代理雨云图片请求的路由。
func RegisterRainImageProxyRoutes(r *gin.Engine) {
	target := "https://cn-nb1.rains3.com"
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("解析目标URL失败: %v\n", err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		// 路径删除 /proxy/rain/api 前缀
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/proxy/rain/image")
		// 添加 Referer
		req.Header.Set("Referer", "https://app.rainyun.com/")
	}
	// 添加 ModifyResponse 来移除上游服务器的 CORS 头
	proxy.ModifyResponse = func(r *http.Response) error {
		// 删除上游服务器的 CORS 相关头部
		r.Header.Del("Access-Control-Allow-Origin")
		r.Header.Del("Access-Control-Allow-Credentials")
		r.Header.Del("Access-Control-Allow-Methods")
		r.Header.Del("Access-Control-Allow-Headers")
		// 对于图片资源，添加必要的安全头
		if strings.HasPrefix(r.Header.Get("Content-Type"), "image/") {
			r.Header.Set("Cross-Origin-Resource-Policy", "cross-origin")
			r.Header.Set("Timing-Allow-Origin", "*")
		}
		//r.Header.Set("Access-Control-Allow-Origin", "*")
		//r.Header.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		return nil
	}

	// 创建反向代理处理函数
	r.Any("/proxy/rain/image/*proxyPath", func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
