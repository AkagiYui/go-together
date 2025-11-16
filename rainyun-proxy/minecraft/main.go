package minecraft

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

func GetMCBE(host string, port int) (*MCBEInfo, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("udp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x24, 0x0D, 0x12, 0xD3, 0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE, 0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78}

	startTime := time.Now()

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)
	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return nil, err
	}

	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	delay := int(time.Since(startTime).Milliseconds())

	response := string(buffer[:n])
	parts := bytes.Split([]byte(response), []byte{';'})

	if len(parts) < 12 {
		return nil, fmt.Errorf("invalid response format")
	}

	return &MCBEInfo{
		Host:            host,
		Port:            port,
		Motd:            string(parts[1]),
		ProtocolVersion: parseInt(parts[2]),
		ServerVersion:   string(parts[3]),
		CurrentPlayers:  parseInt(parts[4]),
		MaxPlayers:      parseInt(parts[5]),
		UniqueID:        string(parts[6]),
		WorldName:       string(parts[7]),
		GameMode:        string(parts[8]),
		PortIPv4:        parseInt(parts[10]),
		PortIPv6:        parseInt(parts[11]),
		Delay:           delay,
	}, nil
}

func parseInt(b []byte) int {
	val := 0
	fmt.Sscan(string(b), &val)
	return val
}
func packVarint(val int) []byte {
	var buf bytes.Buffer
	for i := 0; i < 5; i++ {
		b := val & 0x7F
		val >>= 7
		if val != 0 {
			b |= 0x80
		}
		buf.WriteByte(byte(b))
		if val == 0 {
			break
		}
	}
	return buf.Bytes()
}

func readVarint(conn net.Conn) (int, error) {
	var result int
	var position int

	for i := 0; i < 5; i++ {
		b := make([]byte, 1)
		_, err := conn.Read(b)
		if err != nil {
			return 0, err
		}

		result |= (int(b[0]&0x7F) << position)

		if b[0]&0x80 == 0 {
			break
		}

		position += 7

		if position >= 32 {
			return 0, fmt.Errorf("VarInt is too big")
		}
	}

	return result, nil
}

func GetMCJE(host string, port int) (*MCJEInfo, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Prepare handshake packet
	handshake := bytes.Buffer{}
	handshake.Write([]byte{0x00})   // Packet ID
	handshake.Write(packVarint(47)) // Protocol version (使用固定版本，增加兼容性)
	handshake.Write(packVarint(len(host)))
	handshake.WriteString(host)
	binary.Write(&handshake, binary.BigEndian, uint16(port))
	handshake.Write([]byte{0x01}) // Next state (1 for status)

	// Send handshake packet
	data := bytes.Buffer{}
	data.Write(packVarint(handshake.Len()))
	data.Write(handshake.Bytes())
	if _, err = conn.Write(data.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to send handshake: %v", err)
	}

	// Send status request
	if _, err = conn.Write([]byte{0x01, 0x00}); err != nil {
		return nil, fmt.Errorf("failed to send status request: %v", err)
	}

	// Read packet length
	_, err = readVarint(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %v", err)
	}

	// Read packet ID
	_, err = readVarint(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet ID: %v", err)
	}

	// Read response length
	respLength, err := readVarint(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read response length: %v", err)
	}

	// Read response data
	respData := make([]byte, respLength)
	_, err = io.ReadFull(conn, respData)
	if err != nil {
		return nil, fmt.Errorf("failed to read response data: %v", err)
	}
	fmt.Printf("Raw Response: %s\n", string(respData))
	// Parse JSON response
	var mcResp MCJEResponse
	if err = json.Unmarshal(respData, &mcResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Send and receive ping
	startTime := time.Now()
	pingPacket := bytes.Buffer{}
	pingPacket.Write([]byte{0x01}) // Ping packet ID
	binary.Write(&pingPacket, binary.BigEndian, time.Now().UnixNano()/int64(time.Millisecond))

	data.Reset()
	data.Write(packVarint(pingPacket.Len()))
	data.Write(pingPacket.Bytes())

	if _, err = conn.Write(data.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to send ping: %v", err)
	}

	// Skip ping response packet length and ID
	if _, err = readVarint(conn); err != nil {
		return nil, fmt.Errorf("failed to read ping response length: %v", err)
	}
	if _, err = readVarint(conn); err != nil {
		return nil, fmt.Errorf("failed to read ping response ID: %v", err)
	}

	delay := int(time.Since(startTime).Milliseconds())

	return &MCJEInfo{
		Host:            host,
		Port:            port,
		Description:     mcResp.Description.Text,
		ProtocolVersion: mcResp.Version.Protocol,
		ServerVersion:   mcResp.Version.Name,
		CurrentPlayers:  mcResp.Players.Online,
		MaxPlayers:      mcResp.Players.Max,
		Delay:           delay,
	}, nil
}
