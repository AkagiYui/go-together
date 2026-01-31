# It's MyGO!!!!!

我讨厌开发时代码生成！

go-together 是一个基于 Go 语言开发的多模块项目集合，包含多个独立但相互关联的应用程序和工具库。
项目采用 Go 1.24.5 版本，使用 Go Workspace 管理多个子模块。

## 核心库

### [rest](./rest/) - RESTful API 框架

```shell
go get github.com/akagiyui/go-together/rest@latest
```

[![Go Reference](https://pkg.go.dev/badge/github.com/akagiyui/go-together/rest.svg)](https://pkg.go.dev/github.com/akagiyui/go-together/rest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22-blue.svg)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/akagiyui/go-together/rest)](https://goreportcard.com/report/github.com/akagiyui/go-together/rest)

基于 Go 1.22 ServeMux 的轻量级 RESTful API 框架，支持但不限于：
- 自动参数绑定
- 中间件支持
- 路由组功能
- 数据验证集成

### [common](./common/) - 通用工具库

```shell
go get github.com/akagiyui/go-together/common@latest
```

零依赖库，提供通用工具函数和数据结构，包括但不限于：
- 枚举注册器（EnumRegistry）
- 线程安全的缓存实现
- 数据验证工具
- 加密工具
- 通用响应模型


## 应用程序

### (WIP)[nottodo](./nottodo/) - 待办事项管理

[![Build Status](https://github.com/akagiyui/go-together/actions/workflows/nottodo-build.yml/badge.svg)](https://github.com/AkagiYui/go-together/actions/workflows/nottodo-build.yml)

基于 rest 框架构建的待办事项管理系统。

### (WIP)[arima](./arima/) - 音乐数据库
音频文件管理服务，提供：
- 音频文件上传、下载和列表功能
- S3 对象存储集成
- FFmpeg 音频处理
- 管理接口

### (WIP)[rtsp2s3](./rtsp2s3/) - RTSP 视频录制上传系统
从 RTSP 流录制视频并上传到 S3 的系统，支持：
- 多摄像头同时录制
- 自动分段录制和上传
- 磁盘空间管理
- HTTP 代理支持

### (WIP)[rainyun-proxy](./rainyun-proxy/) - 雨云 API 代理服务
代理雨云 API 和相关服务，包括：
- 雨云 API 和图片服务代理
- Minecraft 服务器管理
- TCP/UDP 客户端代理
- RCON 客户端支持

### [docker-deploy-webhook](./docker-deploy-webhook/) - Docker 部署 Webhook 服务
自动化部署工具，提供：
- GitHub workflow_run 事件接收
- Docker Compose 自动部署
- 多实例部署配置
- HMAC SHA256 签名验证

### [bluestacks](./bluestacks/) - BlueStacks 实例管理工具
BlueStacks Air 安卓模拟器管理工具，功能包括：
- 实例状态监控
- ADB 连接状态显示
- 实时刷新的表格界面
