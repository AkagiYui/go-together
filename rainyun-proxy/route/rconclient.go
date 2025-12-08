package route

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/akagiyui/go-together/rainyun-proxy/rcon"
)

// RCONCommand 前端发送的命令结构
type RCONCommand struct {
	Command string `json:"command"` // 命令类型: exec, close
	Message string `json:"message"` // 具体的RCON命令
}

// RCONMessage RCON消息结构
type RCONMessage struct {
	Type    string `json:"type"`    // 消息类型: message, error
	Message string `json:"message"` // 消息内容
}

// RegisterRCONClientRoutes 注册 RCON 客户端 WebSocket 连接的路由。
func RegisterRCONClientRoutes(r *gin.Engine) {
	r.GET("/rcon/:host/:port/ws", handleRCONWebSocket)
}

// handleRCONWebSocket 处理WebSocket连接
func handleRCONWebSocket(c *gin.Context) {
	// 获取RCON服务器地址和端口参数
	rconHost := c.Param("host")
	rconPort, err := strconv.Atoi(c.Param("port"))
	if err != nil {
		log.Printf("Invalid port number: %v", err)
		return
	}

	// 获取密码参数
	password := c.Query("password")
	if password == "" {
		log.Printf("Password is required")
		return
	}

	// 升级HTTP连接为WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// 创建RCON连接
	rconClient := rcon.NewRCONConnection(rconHost, rconPort, password, 3, 1) // 3次重试，1秒延迟
	err = rconClient.Connect()
	if err != nil {
		log.Printf("RCON connection failed: %v", err)
		ws.WriteJSON(RCONMessage{
			Type:    "error",
			Message: "RCON connection failed: " + err.Error(),
		})
		return
	}
	defer rconClient.Close()

	// 向客户端发送连接成功消息
	ws.WriteJSON(RCONMessage{
		Type:    "message",
		Message: "RCON connection established",
	})

	// 处理从WebSocket接收命令的主循环
	for {
		var cmd RCONCommand
		err := ws.ReadJSON(&cmd)
		if err != nil {
			log.Printf("WebSocket connection closed: %v", err)
			rconClient.Close() // 主动关闭RCON连接
			break
		}

		switch cmd.Command {
		case "exec":
			// 执行RCON命令
			response, err := rconClient.ExecCommand(cmd.Message)
			if err != nil {
				log.Printf("Error executing RCON command: %v", err)
				ws.WriteJSON(RCONMessage{
					Type:    "error",
					Message: "Failed to execute RCON command: " + err.Error(),
				})
				continue
			}

			// 发送命令执行结果
			ws.WriteJSON(RCONMessage{
				Type:    "message",
				Message: response,
			})

		case "close":
			// 关闭连接
			rconClient.Close()
			return

		default:
			ws.WriteJSON(RCONMessage{
				Type:    "error",
				Message: "Unknown command",
			})
		}
	}

	log.Printf("Connection cleanup completed")
}
