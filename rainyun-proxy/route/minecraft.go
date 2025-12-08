// Package route 提供各种代理服务的 HTTP 路由处理器。
package route

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/akagiyui/go-together/rainyun-proxy/minecraft"
)

// RegisterMinecraftRoutes 注册 Minecraft 服务器状态查询的路由。
func RegisterMinecraftRoutes(r *gin.Engine) {
	r.GET("/api/mcbe/:address", handleMCBE)
	r.GET("/api/mcje/:address", handleMCJE)
}

func parseAddress(address string, defaultPort int) (string, int, error) {
	parts := strings.Split(address, ":")

	if len(parts) > 2 {
		return "", 0, fmt.Errorf("invalid address format")
	}

	host := parts[0]
	port := defaultPort

	if len(parts) == 2 {
		parsedPort, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port")
		}
		port = parsedPort
	}

	return host, port, nil
}

func handleMCBE(c *gin.Context) {
	address := c.Param("address")
	host, port, err := parseAddress(address, 19132)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := minecraft.GetMCBE(host, port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

func handleMCJE(c *gin.Context) {
	address := c.Param("address")
	host, port, err := parseAddress(address, 25565)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	info, err := minecraft.GetMCJE(host, port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}
