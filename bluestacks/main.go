package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

// InstanceInfo 存储 BlueStacks 实例信息
type InstanceInfo struct {
	AdbPort     string
	DisplayName string
}

// AdbDevice 存储 ADB 设备信息
type AdbDevice struct {
	Address string
	Status  string
}

// readBlueStacksConfig 读取 BlueStacks 配置文件
// 返回一个 map，键为实例名，值为包含 AdbPort 和 DisplayName 的结构体
func readBlueStacksConfig(configPath string) map[string]InstanceInfo {
	instances := make(map[string]InstanceInfo)

	file, err := os.Open(configPath)
	if err != nil {
		// 如果文件不存在，返回空 map
		return instances
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 去除值的引号
		value = strings.Trim(value, `"`)

		// 解析键的结构：bst.instance.<实例名>.<属性>
		keyParts := strings.Split(key, ".")
		if len(keyParts) < 4 || keyParts[0] != "bst" || keyParts[1] != "instance" {
			continue
		}

		instanceName := keyParts[2]
		propertyPath := strings.Join(keyParts[3:], ".")

		// 获取或创建实例信息
		info := instances[instanceName]

		// 根据属性路径设置值
		if propertyPath == "status.adb_port" {
			info.AdbPort = value
		} else if propertyPath == "display_name" {
			info.DisplayName = value
		}

		instances[instanceName] = info
	}

	return instances
}

// getAdbDevices 执行 adb devices 命令并返回设备列表
func getAdbDevices() []AdbDevice {
	devices := []AdbDevice{}

	cmd := exec.Command("adb", "devices")
	output, err := cmd.Output()
	if err != nil {
		// 如果执行失败，返回空列表
		return devices
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		// 跳过第一行标题行
		if i == 0 {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析设备信息（地址和状态用空格或制表符分隔）
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			devices = append(devices, AdbDevice{
				Address: fields[0],
				Status:  fields[1],
			})
		}
	}

	return devices
}

// getAdbStatus 根据 ADB 端口获取设备状态
func getAdbStatus(adbPort string, devices []AdbDevice) string {
	targetAddress := "127.0.0.1:" + adbPort
	for _, device := range devices {
		if device.Address == targetAddress {
			return device.Status
		}
	}
	return ""
}

// printTable 打印表格
func printTable(instances map[string]InstanceInfo, devices []AdbDevice) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"InstanceName", "DisplayName", "AdbPort", "AdbStatus"})

	// 获取所有实例名并排序
	instanceNames := make([]string, 0, len(instances))
	for instanceName := range instances {
		instanceNames = append(instanceNames, instanceName)
	}
	sort.Strings(instanceNames)

	// 按排序后的顺序添加行
	for _, instanceName := range instanceNames {
		info := instances[instanceName]
		adbStatus := getAdbStatus(info.AdbPort, devices)
		t.AppendRow(table.Row{
			instanceName,
			info.DisplayName,
			info.AdbPort,
			adbStatus,
		})
	}

	t.Render()
}

// clearScreen 清屏并移动光标到开头
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func main() {
	configPath := "/Users/Shared/Library/Application Support/BlueStacks/bluestacks.conf"

	// 创建定时器，每秒执行一次
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	clearScreen()
	instances := readBlueStacksConfig(configPath)
	devices := getAdbDevices()
	printTable(instances, devices)

	// 定时刷新
	for range ticker.C {
		clearScreen()
		instances := readBlueStacksConfig(configPath)
		devices := getAdbDevices()
		printTable(instances, devices)
	}
}
