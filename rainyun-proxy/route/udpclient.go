package route

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// UDPCommand 前端发送的命令结构
type UDPCommand struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

// UDPMessage UDP消息结构
type UDPMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true // 允许所有跨域请求
	},
}

// RegisterUDPClientRoutes 注册 UDP 客户端 WebSocket 连接的路由。
func RegisterUDPClientRoutes(r *gin.Engine) {
	r.GET("/udp/:host/:port/ws", handleUDPWebSocket)
}

// handleUDPWebSocket 处理WebSocket连接
func handleUDPWebSocket(c *gin.Context) {
	// 获取UDP地址和端口参数
	udpHost := c.Param("host")
	udpPort, err := strconv.Atoi(c.Param("port"))
	if err != nil {
		log.Printf("Invalid port number: %v", err)
		return
	}

	// 升级HTTP连接为WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// 创建UDP连接
	udpConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(udpHost),
		Port: udpPort,
	})
	if err != nil {
		log.Printf("UDP connection failed: %v", err)
		ws.WriteJSON(UDPMessage{
			Type:    "error",
			Message: "UDP connection failed: " + err.Error(),
		})
		return
	}
	defer udpConn.Close()

	// 创建停止信号通道
	done := make(chan struct{})
	var wg sync.WaitGroup

	// 处理从UDP接收消息的goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, 1024)
		for {
			select {
			case <-done:
				return
			default:
				n, _, err := udpConn.ReadFromUDP(buffer)
				if err != nil {
					log.Printf("Error reading from UDP: %v", err)
					continue
				}

				// 发送UDP消息到WebSocket
				message := UDPMessage{
					Type:    "message",
					Message: string(buffer[:n]),
				}
				if err := ws.WriteJSON(message); err != nil {
					log.Printf("Error writing to WebSocket: %v", err)
					return
				}
			}
		}
	}()

	// 处理从WebSocket接收命令的主循环
	for {
		var cmd UDPCommand
		err := ws.ReadJSON(&cmd)
		if err != nil {
			log.Printf("Error reading from WebSocket: %v", err)
			break
		}

		switch cmd.Command {
		case "send":
			// 发送消息到UDP服务器
			_, err := udpConn.Write([]byte(cmd.Message))
			if err != nil {
				log.Printf("Error writing to UDP: %v", err)
				ws.WriteJSON(UDPMessage{
					Type:    "error",
					Message: "Failed to send UDP message: " + err.Error(),
				})
			}
		case "close":
			// 关闭连接
			close(done)
			return
		default:
			ws.WriteJSON(UDPMessage{
				Type:    "error",
				Message: "Unknown command",
			})
		}
	}

	// 关闭done通道并等待goroutine结束
	close(done)
	wg.Wait()
}
