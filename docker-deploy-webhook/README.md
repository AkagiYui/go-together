# Docker Deploy Webhook

一个用于接收 GitHub workflow_run 事件并自动执行 Docker Compose 部署的 webhook 服务。

## 功能特性

- ✅ 接收 GitHub workflow_run 完成事件
- ✅ 验证 GitHub webhook 签名（HMAC SHA256）
- ✅ 根据配置规则匹配仓库、分支和工作流
- ✅ 自动执行 Docker Compose 部署流程
- ✅ 完整的日志记录
- ✅ 健康检查端点

## 环境变量配置

所有配置通过环境变量提供：

| 环境变量 | 必需 | 默认值 | 说明 |
|---------|------|--------|------|
| `PORT` | 否 | `8080` | HTTP 服务监听端口 |
| `WEBHOOK_SECRET` | 是 | - | GitHub webhook secret，用于验证请求 |
| `REPOSITORY_NAME` | 是 | - | 要匹配的仓库名称（格式：owner/repo） |
| `BRANCH_NAME` | 是 | - | 要匹配的分支名称 |
| `WORKFLOW_FILE_NAME` | 是 | - | 要匹配的工作流文件名 |
| `COMPOSE_FILE_PATH` | 否 | `docker-compose.yml` | Docker Compose 文件路径 |
| `COMPOSE_PROJECT_DIR` | 否 | `.` | Docker Compose 项目目录 |

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

### 2. 配置环境变量

复制示例配置文件并修改：

```bash
cp .env.example .env
# 编辑 .env 文件，填入实际配置
```

### 3. 运行

```bash
# 加载环境变量并运行
export $(cat .env | xargs)
./webhook-server
```

或者直接设置环境变量：

```bash
PORT=8080 \
WEBHOOK_SECRET=your_secret \
REPOSITORY_NAME=owner/repo \
BRANCH_NAME=main \
WORKFLOW_FILE_NAME=deploy.yml \
COMPOSE_FILE_PATH=/path/to/docker-compose.yml \
COMPOSE_PROJECT_DIR=/path/to/project \
./webhook-server
```

## GitHub Webhook 配置

1. 进入 GitHub 仓库的 Settings → Webhooks → Add webhook
2. 配置以下内容：
   - **Payload URL**: `http://your-server:8080/webhook`
   - **Content type**: `application/json`
   - **Secret**: 与 `WEBHOOK_SECRET` 环境变量相同的值
   - **Which events**: 选择 "Let me select individual events"，勾选 "Workflow runs"
   - **Active**: 勾选

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
5. 匹配配置规则：
   - 仓库名称匹配
   - 分支名称匹配
   - 工作流文件名匹配
   - 工作流状态为 `completed`
   - 工作流结论为 `success`
6. 如果匹配成功，执行部署：
   - `docker compose pull` - 拉取最新镜像
   - `docker compose down` - 停止并删除现有容器
   - `docker compose up -d` - 重新创建并启动容器

## 日志示例

```
2024/01/01 12:00:00 Starting webhook server on port 8080
2024/01/01 12:00:00 Monitoring repository: owner/repo
2024/01/01 12:00:00 Monitoring branch: main
2024/01/01 12:00:00 Monitoring workflow: deploy.yml
2024/01/01 12:00:00 Docker Compose file: docker-compose.yml
2024/01/01 12:00:00 Docker Compose project directory: /path/to/project
2024/01/01 12:00:00 Server listening on :8080
2024/01/01 12:01:00 Received workflow_run event: action=completed, repo=owner/repo, branch=main, workflow=.github/workflows/deploy.yml, status=completed, conclusion=success
2024/01/01 12:01:00 Conditions matched, starting deployment...
2024/01/01 12:01:00 Starting Docker Compose deployment...
2024/01/01 12:01:00 Pulling latest images...
2024/01/01 12:01:00 Executing: docker compose -f docker-compose.yml pull
2024/01/01 12:01:05 Stopping and removing existing containers...
2024/01/01 12:01:05 Executing: docker compose -f docker-compose.yml down
2024/01/01 12:01:10 Starting containers...
2024/01/01 12:01:10 Executing: docker compose -f docker-compose.yml up -d
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
WorkingDirectory=/opt/webhook
ExecStart=/opt/webhook/webhook-server

# 环境变量配置
Environment="PORT=8080"
Environment="WEBHOOK_SECRET=your_webhook_secret_here"
Environment="REPOSITORY_NAME=owner/repo"
Environment="BRANCH_NAME=main"
Environment="WORKFLOW_FILE_NAME=deploy.yml"
Environment="COMPOSE_FILE_PATH=/path/to/docker-compose.yml"
Environment="COMPOSE_PROJECT_DIR=/path/to/project"

# 或者从文件加载环境变量
# EnvironmentFile=/opt/webhook/.env

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

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable webhook
sudo systemctl start webhook
sudo systemctl status webhook
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

1. **使用强 webhook secret**: 确保 `WEBHOOK_SECRET` 足够复杂且随机（建议使用 `openssl rand -hex 32` 生成）
2. **HTTPS**: 在生产环境中使用反向代理（如 Nginx）提供 HTTPS
3. **防火墙**: 限制只允许 GitHub 的 IP 地址访问 webhook 端点
4. **权限控制**: 确保运行服务的用户有执行 docker compose 命令的权限
5. **日志监控**: 定期检查日志，监控异常活动

## 故障排查

### 签名验证失败

- 检查 `WEBHOOK_SECRET` 是否与 GitHub webhook 配置中的 secret 一致
- 确认 GitHub webhook 配置中选择了正确的 Content type（application/json）

### 条件不匹配

- 检查日志中的详细匹配信息
- 确认环境变量配置正确
- 注意工作流文件名是检查路径后缀，例如 `.github/workflows/deploy.yml` 会匹配 `deploy.yml`

### 部署失败

- 检查 Docker 是否正常运行
- 确认运行服务的用户有 Docker 权限
- 检查 `COMPOSE_FILE_PATH` 和 `COMPOSE_PROJECT_DIR` 是否正确
- 查看 docker compose 命令的输出日志

## 许可证

MIT

