package route

import (
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// TCPCommand 前端发送的命令结构
type TCPCommand struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

// TCPMessage TCP消息结构
type TCPMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func RegisterTCPClientRoutes(r *gin.Engine) {
	r.GET("/tcp/:host/:port/ws", handleTCPWebSocket)
}

// handleTCPWebSocket 处理WebSocket连接
func handleTCPWebSocket(c *gin.Context) {
	// 获取TCP地址和端口参数
	tcpHost := c.Param("host")
	tcpPort, err := strconv.Atoi(c.Param("port"))
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

	// 创建TCP连接
	tcpAddr := net.TCPAddr{
		IP:   net.ParseIP(tcpHost),
		Port: tcpPort,
	}
	tcpConn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err != nil {
		log.Printf("TCP connection failed: %v", err)
		ws.WriteJSON(TCPMessage{
			Type:    "error",
			Message: "TCP connection failed: " + err.Error(),
		})
		return
	}
	defer tcpConn.Close()

	// 创建停止信号通道
	done := make(chan struct{})
	var wg sync.WaitGroup

	// 处理从TCP接收消息的goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, 1024)
		for {
			select {
			case <-done:
				return
			default:
				// 设置读取超时
				//tcpConn.SetReadDeadline(time.Now().Add(time.Second * 5))

				n, err := tcpConn.Read(buffer)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// 超时，继续下一次读取
						continue
					}
					log.Printf("Error reading from TCP: %v", err)
					ws.WriteJSON(TCPMessage{
						Type:    "error",
						Message: "TCP read error: " + err.Error(),
					})
					ws.Close() // 主动关闭WebSocket连接
					return
				}

				// 发送TCP消息到WebSocket
				message := TCPMessage{
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
		var cmd TCPCommand
		err := ws.ReadJSON(&cmd)
		if err != nil {
			log.Printf("WebSocket connection closed: %v", err)
			// 主动关闭TCP连接
			tcpConn.Close()
			break
		}

		switch cmd.Command {
		case "send":
			// 发送消息到TCP服务器
			_, err := tcpConn.Write([]byte(cmd.Message))
			if err != nil {
				log.Printf("Error writing to TCP: %v", err)
				ws.WriteJSON(TCPMessage{
					Type:    "error",
					Message: "Failed to send TCP message: " + err.Error(),
				})
			}
		case "close":
			// 关闭连接
			close(done)
			return
		default:
			ws.WriteJSON(TCPMessage{
				Type:    "error",
				Message: "Unknown command",
			})
		}
	}

	// 关闭done通道并等待goroutine结束
	close(done)
	wg.Wait()
}
