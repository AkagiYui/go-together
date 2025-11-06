// Package main implements a GitHub webhook server that listens for workflow_run events
// and automatically deploys applications using Docker Compose.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Config 存储应用配置
type Config struct {
	Port              string // HTTP 服务监听端口
	WebhookSecret     string // GitHub webhook secret
	RepositoryName    string // 要匹配的仓库名称
	BranchName        string // 要匹配的分支名称
	WorkflowFileName  string // 要匹配的工作流文件名
	ComposeFilePath   string // Docker Compose 文件路径
	ComposeProjectDir string // Docker Compose 项目目录
}

// WorkflowRunPayload GitHub workflow_run 事件的 payload 结构
type WorkflowRunPayload struct {
	Action      string `json:"action"`
	WorkflowRun struct {
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		HeadBranch string `json:"head_branch"`
		Path       string `json:"path"`
	} `json:"workflow_run"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// loadConfig 从环境变量加载配置
func loadConfig() (*Config, error) {
	config := &Config{
		Port:              getEnv("PORT", "8080"),
		WebhookSecret:     getEnv("WEBHOOK_SECRET", ""),
		RepositoryName:    getEnv("REPOSITORY_NAME", ""),
		BranchName:        getEnv("BRANCH_NAME", ""),
		WorkflowFileName:  getEnv("WORKFLOW_FILE_NAME", ""),
		ComposeFilePath:   getEnv("COMPOSE_FILE_PATH", "docker-compose.yml"),
		ComposeProjectDir: getEnv("COMPOSE_PROJECT_DIR", "."),
	}

	// 验证必需的配置
	if config.WebhookSecret == "" {
		return nil, fmt.Errorf("WEBHOOK_SECRET is required")
	}
	if config.RepositoryName == "" {
		return nil, fmt.Errorf("REPOSITORY_NAME is required")
	}
	if config.BranchName == "" {
		return nil, fmt.Errorf("BRANCH_NAME is required")
	}
	if config.WorkflowFileName == "" {
		return nil, fmt.Errorf("WORKFLOW_FILE_NAME is required")
	}

	return config, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// verifySignature 验证 GitHub webhook 签名
func verifySignature(secret string, signature string, body []byte) bool {
	if signature == "" {
		return false
	}

	// GitHub 签名格式: sha256=<hash>
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	// 提取签名哈希值
	expectedHash := signature[7:]

	// 计算 HMAC SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actualHash := hex.EncodeToString(mac.Sum(nil))

	// 比较签名
	return hmac.Equal([]byte(expectedHash), []byte(actualHash))
}

// matchesConditions 检查 payload 是否匹配配置的条件
func matchesConditions(payload *WorkflowRunPayload, config *Config) bool {
	// 检查仓库名称
	if payload.Repository.FullName != config.RepositoryName {
		log.Printf("Repository mismatch: got %s, expected %s", payload.Repository.FullName, config.RepositoryName)
		return false
	}

	// 检查分支名称
	if payload.WorkflowRun.HeadBranch != config.BranchName {
		log.Printf("Branch mismatch: got %s, expected %s", payload.WorkflowRun.HeadBranch, config.BranchName)
		return false
	}

	// 检查工作流文件名
	if !strings.HasSuffix(payload.WorkflowRun.Path, config.WorkflowFileName) {
		log.Printf("Workflow file mismatch: got %s, expected suffix %s", payload.WorkflowRun.Path, config.WorkflowFileName)
		return false
	}

	// 检查工作流是否完成且成功
	if payload.WorkflowRun.Status != "completed" {
		log.Printf("Workflow not completed: status is %s", payload.WorkflowRun.Status)
		return false
	}

	if payload.WorkflowRun.Conclusion != "success" {
		log.Printf("Workflow not successful: conclusion is %s", payload.WorkflowRun.Conclusion)
		return false
	}

	return true
}

// executeDockerCompose 执行 Docker Compose 部署命令
func executeDockerCompose(config *Config) error {
	log.Println("Starting Docker Compose deployment...")

	// 1. 拉取最新镜像
	log.Println("Pulling latest images...")
	if err := runDockerComposeCommand(config, "pull"); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	// 2. 停止并删除现有容器
	log.Println("Stopping and removing existing containers...")
	if err := runDockerComposeCommand(config, "down"); err != nil {
		return fmt.Errorf("failed to stop containers: %w", err)
	}

	// 3. 重新创建并启动容器
	log.Println("Starting containers...")
	if err := runDockerComposeCommand(config, "up", "-d"); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}

	log.Println("Docker Compose deployment completed successfully")
	return nil
}

// runDockerComposeCommand 运行 docker compose 命令
func runDockerComposeCommand(config *Config, args ...string) error {
	// 构建命令参数
	cmdArgs := []string{"compose", "-f", config.ComposeFilePath}
	cmdArgs = append(cmdArgs, args...)

	// 创建命令
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Dir = config.ComposeProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Executing: docker %s", strings.Join(cmdArgs, " "))

	// 执行命令
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// webhookHandler 处理 webhook 请求
func webhookHandler(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 只接受 POST 请求
		if r.Method != http.MethodPost {
			log.Printf("Invalid method: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 验证签名
		signature := r.Header.Get("X-Hub-Signature-256")
		if !verifySignature(config.WebhookSecret, signature, body) {
			log.Println("Invalid signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		// 检查事件类型
		eventType := r.Header.Get("X-GitHub-Event")
		if eventType != "workflow_run" {
			log.Printf("Ignoring event type: %s", eventType)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Event type %s ignored", eventType)
			return
		}

		// 解析 payload
		var payload WorkflowRunPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("Failed to parse payload: %v", err)
			http.Error(w, "Failed to parse payload", http.StatusBadRequest)
			return
		}

		log.Printf("Received workflow_run event: action=%s, repo=%s, branch=%s, workflow=%s, status=%s, conclusion=%s",
			payload.Action,
			payload.Repository.FullName,
			payload.WorkflowRun.HeadBranch,
			payload.WorkflowRun.Path,
			payload.WorkflowRun.Status,
			payload.WorkflowRun.Conclusion,
		)

		// 检查是否匹配条件
		if !matchesConditions(&payload, config) {
			log.Println("Conditions not matched, skipping deployment")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Conditions not matched")
			return
		}

		// 执行部署
		log.Println("Conditions matched, starting deployment...")
		if err := executeDockerCompose(config); err != nil {
			log.Printf("Deployment failed: %v", err)
			http.Error(w, "Deployment failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Deployment successful")
	}
}

// healthHandler 健康检查处理器
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting webhook server on port %s", config.Port)
	log.Printf("Monitoring repository: %s", config.RepositoryName)
	log.Printf("Monitoring branch: %s", config.BranchName)
	log.Printf("Monitoring workflow: %s", config.WorkflowFileName)
	log.Printf("Docker Compose file: %s", config.ComposeFilePath)
	log.Printf("Docker Compose project directory: %s", config.ComposeProjectDir)

	// 设置路由
	http.HandleFunc("/webhook", webhookHandler(config))
	http.HandleFunc("/health", healthHandler)

	// 启动服务器
	addr := ":" + config.Port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
