# keep-traffic-ratio

一个流量比例保持工具，用于监控网络流量并通过自动下载来维持指定的下载/上传流量比例。

## 功能特性

- 📊 **实时流量监控** - 监控指定网络接口的下载和上传流量
- 🔄 **自动比例调节** - 当下载流量比例低于目标值时自动下载文件
- 🎨 **终端 UI** - 美观的终端用户界面，显示实时流量统计
- 🖥️ **无 UI 模式** - 支持纯文本输出模式，适合脚本或后台运行
- 🧪 **试运行模式** - Dry-run 模式，模拟下载而不实际消耗流量
- 🌐 **自动获取资源** - 从 lolicp.com 自动获取可用的流量资源

## 安装

### 从源码构建

```bash
git clone https://github.com/akagiyui/go-together.git
cd go-together/keep-traffic-ratio
go build .
```

### 前置要求

- Go 1.26 或更高版本
- 支持 Linux、macOS、Windows 系统

## 使用方法

### 基本用法

```bash
./keep-traffic-ratio -interface <网络接口名> -ratio <目标比例>
```

### 参数说明

| 参数         | 说明                        | 是否必需 | 默认值 |
| ------------ | --------------------------- | -------- | ------ |
| `-interface` | 要监控的网络接口名称        | 是       | 无     |
| `-ratio`     | 目标下载流量比例（0.0-1.0） | 是       | 无     |
| `-dryrun`    | 试运行模式，不实际下载文件  | 否       | false  |
| `-no-ui`     | 禁用终端 UI，使用标准输出   | 否       | false  |

### 示例

#### 基本使用

```bash
# 监控 eth0 接口，保持下载流量至少占总流量的 50%
./keep-traffic-ratio -interface eth0 -ratio 0.5
```

#### 试运行模式

```bash
# 试运行模式，模拟下载而不实际消耗流量
./keep-traffic-ratio -interface eth0 -ratio 0.5 -dryrun
```

#### 无 UI 模式

```bash
# 适合在脚本中使用或后台运行
./keep-traffic-ratio -interface eth0 -ratio 0.5 -no-ui
```

#### 完整参数

```bash
# 组合使用所有参数
./keep-traffic-ratio -interface eth0 -ratio 0.8 -dryrun -no-ui
```

## 工作原理

1. **流量监控**：使用 `gopsutil` 库实时监控指定网络接口的下载和上传流量
2. **比例计算**：计算当前下载流量占总流量（下载+上传）的比例
3. **自动下载**：当下载比例低于目标值时，从 lolicp.com 获取流量资源 URL 并随机选择一个进行下载
4. **持续运行**：每秒更新一次流量统计，确保流量比例维持在目标水平

## 查看可用网络接口

如果不确定可用的网络接口名称，可以在运行时查看错误提示：

```bash
./keep-traffic-ratio -interface invalid -ratio 0.5
```

程序会列出所有可用的网络接口。

## 输出说明

### 终端 UI 模式

终端 UI 模式会显示：
- 网络接口名称
- 当前下载/上传流量
- 当前下载/上传比例
- 目标比例
- 下载状态和进度
- 操作日志

### 无 UI 模式

无 UI 模式会在标准输出中显示：
```
Interface: eth0 | Download: 1.5 GB (65.23%) | 保持至少: 50.00% | Upload: 800 MB | Total: 1.5 GB / 800 MB | Downloading: false
```

## 依赖

- [gopsutil](https://github.com/shirou/gopsutil) - 系统和进程监控库
- [BubbleTea](https://charm.land/bubbletea/v2) - 终端 UI 框架
- [lipgloss](https://charm.land/lipgloss/v2) - 终端样式库
- [goquery](https://github.com/PuerkitoBio/goquery) - HTML 解析库
- [bubblezone](https://github.com/lrstanley/bubblezone) - BubbleTea 区域管理

## 使用场景

- 🌐 **ISP 流量计费优化** - 某些 ISP 按下载流量计费，上传不计费，可通过保持下载比例来优化
- 📊 **流量平衡** - 保持下载和上传流量的平衡
- 🧪 **测试和调试** - 使用 dry-run 模式测试流量监控系统

## 注意事项

- 需要足够的系统权限来监控网络接口
- 确保网络接口名称正确（不同系统可能有不同的命名规则）
- 下载的流量会实际消耗网络带宽
- 在 dry-run 模式下不会实际下载文件
