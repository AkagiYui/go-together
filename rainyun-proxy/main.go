// Package main 提供雨云 API 和各种网络服务的代理服务器。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/akagiyui/go-together/rainyun-proxy/route"
)

func healthcheck() {
	log.Println("Health check")
	ptr := flag.Bool("healthcheck", false, "Health check")
	urlPtr := flag.String("url", "http://127.0.0.1:28183", "URL to health check")
	timeoutPtr := flag.Int("timeout", 5, "Timeout in seconds")

	flag.Parse()
	if *ptr != true {
		return
	}

	client := &http.Client{
		Timeout: time.Duration(*timeoutPtr) * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("%s/health", *urlPtr))
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != 200 {
		fmt.Printf("Health check failed: status code %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Printf("Health check passed: status code %d\n", resp.StatusCode)
	os.Exit(0)
}

func main() {
	healthcheck()

	r := gin.Default()
	// 配置 CORS 中间件
	config := cors.Config{
		//AllowOrigins:     []string{"https://rya.akagiyui.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "x-api-key"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12小时的预检请求缓存
	}
	// 添加自定义源验证函数，允许所有 localhost 源
	config.AllowOriginFunc = func(origin string) bool {
		return strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "https://localhost:") ||
			origin == "https://rya.akagiyui.com"
	}

	r.Use(cors.New(config))

	route.RegisterRainAPIProxyRoutes(r)
	route.RegisterRainImageProxyRoutes(r)
	route.RegisterMinecraftRoutes(r)
	route.RegisterUDPClientRoutes(r)
	route.RegisterTCPClientRoutes(r)
	route.RegisterRCONClientRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	err := r.Run(":28183")
	if err != nil {
		log.Fatalf("启动服务失败: %v\n", err)
		return
	}
}
