# Docker Deploy Webhook

一个用于接收 GitHub workflow_run 事件并自动执行 Docker Compose 部署的 webhook 服务。

## 功能特性

- ✅ 接收 GitHub workflow_run 完成事件
- ✅ 验证 GitHub webhook 签名（HMAC SHA256）
- ✅ 支持多实例部署配置
- ✅ 根据配置规则匹配仓库、分支和工作流
- ✅ 自动执行 Docker Compose 部署流程
- ✅ 完整的日志记录
- ✅ 健康检查端点

## 配置文件

使用 TOML 格式的配置文件 `config.toml`（相对于程序运行时的工作目录）。

### 配置结构

```toml
[server]
port = "8080"                    # HTTP 服务监听端口
webhook_secret = "your-secret"   # GitHub webhook secret
log_level = "info"               # 日志级别（可选）

[[instances]]
repository_name = "owner/repo"              # 仓库名称
branch_name = "main"                        # 分支名称
workflow_file_name = "deploy.yml"           # 工作流文件名
compose_file_path = "/path/to/compose.yml"  # Docker Compose 文件路径
compose_project_dir = "/path/to/project"    # Docker Compose 项目目录

# 可以配置多个实例
[[instances]]
repository_name = "owner/another-repo"
branch_name = "production"
workflow_file_name = "prod-deploy.yml"
compose_file_path = "/opt/app/docker-compose.yml"
compose_project_dir = "/opt/app"
```

### 配置说明

**服务器配置 `[server]`**

| 字段 | 必需 | 说明 |
|------|------|------|
| `port` | 是 | HTTP 服务监听端口 |
| `webhook_secret` | 是 | GitHub webhook secret，用于验证请求 |
| `log_level` | 否 | 日志级别（debug/info/warn/error） |

**实例配置 `[[instances]]`**

| 字段 | 必需 | 说明 |
|------|------|------|
| `repository_name` | 是 | 要匹配的仓库名称（格式：owner/repo） |
| `branch_name` | 是 | 要匹配的分支名称 |
| `workflow_file_name` | 是 | 要匹配的工作流文件名 |
| `compose_file_path` | 是 | Docker Compose 文件路径 |
| `compose_project_dir` | 是 | Docker Compose 项目目录 |

## 快速开始

### 1. 获取二进制文件

#### 方式 1: 从 GitHub Actions 下载预编译的二进制文件

访问本项目的 [Actions](../../actions) 页面，选择最新的成功构建，下载对应平台的二进制文件：

- Linux (amd64, arm64, armv7, armv6)
- Windows (amd64, arm64)
- macOS (amd64, arm64)

#### 方式 2: 从源码编译

```bash
cd docker-deploy-webhook
go build -o webhook-server
```

或者交叉编译到其他平台：

```bash
# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o webhook-server-linux-arm64

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o webhook-server-windows-amd64.exe

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o webhook-server-darwin-arm64
```

### 2. 创建配置文件

复制示例配置文件并修改：

```bash
cp config.toml.example config.toml
# 编辑 config.toml 文件，填入实际配置
```

### 3. 运行

```bash
# 确保 config.toml 在当前目录
./webhook-server
```

**注意**：配置文件路径 `./config.toml` 是相对于程序运行时的工作目录，而不是可执行文件所在目录。

## GitHub Webhook 配置

1. 进入 GitHub 仓库的 Settings → Webhooks → Add webhook
2. 配置以下内容：
   - **Payload URL**: `http://your-server:8080/webhook`
   - **Content type**: `application/json`
   - **Secret**: 与配置文件中 `server.webhook_secret` 相同的值
   - **Which events**: 选择 "Let me select individual events"，勾选 "Workflow runs"
   - **Active**: 勾选

**注意**：多个仓库可以使用同一个 webhook 端点，只要它们共享相同的 secret，并在配置文件中定义了对应的实例。

## API 端点

### POST /webhook

接收 GitHub webhook 请求。

**请求头：**
- `X-Hub-Signature-256`: GitHub 签名
- `X-GitHub-Event`: 事件类型（应为 `workflow_run`）

**响应：**
- `200 OK`: 请求处理成功
- `400 Bad Request`: 请求格式错误
- `401 Unauthorized`: 签名验证失败
- `405 Method Not Allowed`: 非 POST 请求
- `500 Internal Server Error`: 部署失败

### GET /health

健康检查端点。

**响应：**
- `200 OK`: 服务正常运行

## 工作流程

1. GitHub workflow 完成后触发 webhook
2. 服务验证请求签名
3. 检查事件类型是否为 `workflow_run`
4. 解析 payload 并提取：
   - 仓库名称
   - 分支名称
   - 工作流文件名
   - 工作流状态和结论
5. 遍历所有配置的实例，查找匹配的实例：
   - 仓库名称匹配
   - 分支名称匹配
   - 工作流文件名匹配
   - 工作流状态为 `completed`
   - 工作流结论为 `success`
6. 如果找到匹配的实例，执行部署：
   - `docker compose pull` - 拉取最新镜像
   - `docker compose down` - 停止并删除现有容器
   - `docker compose up -d` - 重新创建并启动容器

## 日志示例

```
2024/01/01 12:00:00 Starting webhook server on port 8080
2024/01/01 12:00:00 Loaded 2 deployment instance(s):
2024/01/01 12:00:00   [1] Repository: owner/app1, Branch: main, Workflow: deploy.yml
2024/01/01 12:00:00   [2] Repository: owner/app2, Branch: production, Workflow: prod-deploy.yml
2024/01/01 12:00:00 Server listening on :8080
2024/01/01 12:01:00 Received workflow_run event: action=completed, repo=owner/app1, branch=main, workflow=.github/workflows/deploy.yml, status=completed, conclusion=success
2024/01/01 12:01:00 Found matching instance for owner/app1/main, starting deployment...
2024/01/01 12:01:00 Starting Docker Compose deployment for owner/app1...
2024/01/01 12:01:00 Pulling latest images...
2024/01/01 12:01:00 Executing: docker compose -f /opt/app1/docker-compose.yml pull
2024/01/01 12:01:05 Stopping and removing existing containers...
2024/01/01 12:01:05 Executing: docker compose -f /opt/app1/docker-compose.yml down
2024/01/01 12:01:10 Starting containers...
2024/01/01 12:01:10 Executing: docker compose -f /opt/app1/docker-compose.yml up -d
2024/01/01 12:01:15 Docker Compose deployment completed successfully
```

## 部署方式

### systemd 服务（推荐）

创建 systemd 服务文件 `/etc/systemd/system/webhook.service`:

```ini
[Unit]
Description=Docker Deploy Webhook Service
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=root
# 工作目录必须包含 config.toml 文件
WorkingDirectory=/opt/webhook
ExecStart=/opt/webhook/webhook-server

# 重启策略
Restart=always
RestartSec=10

# 日志配置
StandardOutput=journal
StandardError=journal
SyslogIdentifier=webhook

[Install]
WantedBy=multi-user.target
```

部署步骤：

```bash
# 创建部署目录
sudo mkdir -p /opt/webhook

# 复制可执行文件
sudo cp webhook-server /opt/webhook/

# 创建配置文件
sudo cp config.toml.example /opt/webhook/config.toml
sudo nano /opt/webhook/config.toml  # 编辑配置

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable webhook
sudo systemctl start webhook
sudo systemctl status webhook

# 查看日志
sudo journalctl -u webhook -f
```

### 使用反向代理（Nginx + HTTPS）

```nginx
server {
    listen 443 ssl;
    server_name webhook.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location /webhook {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://localhost:8080;
    }
}
```

## 安全建议

1. **使用强 webhook secret**: 确保 `server.webhook_secret` 足够复杂且随机（建议使用 `openssl rand -hex 32` 生成）
2. **HTTPS**: 在生产环境中使用反向代理（如 Nginx）提供 HTTPS
3. **防火墙**: 限制只允许 GitHub 的 IP 地址访问 webhook 端点
4. **权限控制**: 确保运行服务的用户有执行 docker compose 命令的权限
5. **日志监控**: 定期检查日志，监控异常活动
6. **配置文件权限**: 限制 `config.toml` 的读取权限（`chmod 600 config.toml`）

## 故障排查

### 配置文件未找到

- 确认 `config.toml` 文件在程序运行时的工作目录中
- 如果使用 systemd，检查 `WorkingDirectory` 设置
- 配置文件路径是相对于工作目录，不是可执行文件所在目录

### 签名验证失败

- 检查 `server.webhook_secret` 是否与 GitHub webhook 配置中的 secret 一致
- 确认 GitHub webhook 配置中选择了正确的 Content type（application/json）

### 没有匹配的实例

- 检查日志中的详细匹配信息
- 确认配置文件中的实例配置正确
- 注意工作流文件名是检查路径后缀，例如 `.github/workflows/deploy.yml` 会匹配 `deploy.yml`
- 确认仓库名称格式为 `owner/repo`

### 部署失败

- 检查 Docker 是否正常运行
- 确认运行服务的用户有 Docker 权限
- 检查实例配置中的 `compose_file_path` 和 `compose_project_dir` 是否正确
- 查看 docker compose 命令的输出日志

## 许可证

MIT

