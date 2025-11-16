// Package rcon 提供 RCON (远程控制台) 客户端功能，用于连接游戏服务器。
package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

// Connection 表示到远程服务器的 RCON 连接。
type Connection struct {
	conn       net.Conn
	ip         string
	port       int
	password   string
	retryCount int
	retryDelay int
}

// NewRCONConnection 创建一个新的 RCON 连接实例，使用指定的参数。
func NewRCONConnection(ip string, port int, password string, retryCount int, retryDelay int) *Connection {
	connection := &Connection{
		ip:         ip,
		port:       port,
		password:   password,
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
	return connection
}

// Connect 建立到 RCON 服务器的连接并执行身份验证。
func (c *Connection) Connect() error {
	c.Close()
	address := net.JoinHostPort(c.ip, fmt.Sprintf("%d", c.port))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.conn = conn
	// auth
	if len(c.password) > 0 {
		c.auth()
	}
	return nil
}

func (c *Connection) auth() (string, error) {
	packet := createPacket(dumbID, serverDataAuth, c.password)
	_, wRrr := c.conn.Write(packet)
	if wRrr != nil {
		return "", wRrr
	}

	buf := make([]byte, maxPacketSize)
	_, rErr := c.conn.Read(buf)
	if rErr != nil {
		return "", rErr
	}
	pkg, err := readPacket(buf)
	if err != nil {
		return "", errors.New("auth failed, wrong password")
	}

	if pkg.ID != dumbID {
		return "", err
	}

	return pkg.Body, nil
}

// ExecCommand 在 RCON 服务器上执行命令并返回响应。
func (c *Connection) ExecCommand(command string) (string, error) {
	return c.execCommandImp(command, c.retryCount)
}

func (c *Connection) execCommandImp(command string, retryCount int) (string, error) {
	if c.conn == nil {
		if c.retryCount > 0 {
			if c.retryDelay > 0 {
				time.Sleep(time.Duration(c.retryDelay) * time.Second)
			}
			c.Connect()
			return c.execCommandImp(command, retryCount-1)
		}
		return "", fmt.Errorf("rcon connection is not established")
	}
	resp, err := c.execute(command)
	if err != nil {
		if retryCount > 0 {
			if c.retryDelay > 0 {
				time.Sleep(time.Duration(c.retryDelay) * time.Second)
			}
			c.Connect()
			return c.execCommandImp(command, retryCount-1)
		}
		return "", err
	}
	return resp, nil
}

func (c *Connection) execute(command string) (string, error) {
	packet := createPacket(dumbID, serverDataExecCommand, command)
	_, wRrr := c.conn.Write(packet)
	if wRrr != nil {
		return "", wRrr
	}

	buf := make([]byte, maxPacketSize)

	_, rErr := c.conn.Read(buf)
	if rErr != nil {
		return "", rErr
	}

	pkg, err := readPacket(buf)
	if err != nil {
		return "", err
	}
	return pkg.Body, nil
}

// Close 关闭 RCON 连接。
func (c *Connection) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

const (
	serverDataAuth          int32 = 3
	serverDataAuthResponse  int32 = 2
	serverDataExecCommand   int32 = 2
	serverDataResponseValue int32 = 0
	maxPacketSize           int32 = 4096

	dumbID int32 = 0

	headerLength       = 10
	maximumPackageSize = 4096
)

type rconPacket struct {
	Size int32
	ID   int32
	Type int32
	Body string
}

func createPacket(id int32, pkgType int32, command string) []byte {
	commandBytes := []byte(command)
	// id:4  type:4  end:2
	size := int32(10 + len(commandBytes))

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, size)
	binary.Write(&buf, binary.LittleEndian, id)
	binary.Write(&buf, binary.LittleEndian, pkgType)
	buf.Write(commandBytes)
	buf.Write([]byte{0x00, 0x00})
	return buf.Bytes()
}

func readPacket(buf []byte) (rconPacket, error) {
	packet := &rconPacket{}
	packet.Size = int32(binary.LittleEndian.Uint32(buf[0:4]))
	packet.ID = int32(binary.LittleEndian.Uint32(buf[4:8]))
	packet.Type = int32(binary.LittleEndian.Uint32(buf[8:12]))
	bodyLength := packet.Size - 10
	packet.Body = string(buf[12 : 12+bodyLength])

	return *packet, nil
}
