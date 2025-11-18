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

	"github.com/BurntSushi/toml"
)

// Config 存储应用配置
type Config struct {
	Server    ServerConfig     `toml:"server"`    // 服务器配置
	Instances []InstanceConfig `toml:"instances"` // 部署实例配置数组
}

// ServerConfig 服务器全局配置
type ServerConfig struct {
	Port          string `toml:"port"`           // HTTP 服务监听端口
	WebhookSecret string `toml:"webhook_secret"` // GitHub webhook secret
	LogLevel      string `toml:"log_level"`      // 日志级别 (可选)
}

// InstanceConfig 单个部署实例配置
type InstanceConfig struct {
	RepositoryName    string `toml:"repository_name"`     // 要匹配的仓库名称
	BranchName        string `toml:"branch_name"`         // 要匹配的分支名称
	WorkflowFileName  string `toml:"workflow_file_name"`  // 要匹配的工作流文件名
	ComposeFilePath   string `toml:"compose_file_path"`   // Docker Compose 文件路径
	ComposeProjectDir string `toml:"compose_project_dir"` // Docker Compose 项目目录
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

// loadConfig 从 TOML 配置文件加载配置
func loadConfig() (*Config, error) {
	const configPath = "./config.toml"

	// 读取配置文件
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validateConfig 验证配置的有效性
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}
	if config.Server.WebhookSecret == "" {
		return fmt.Errorf("server.webhook_secret is required")
	}

	// 验证至少有一个部署实例
	if len(config.Instances) == 0 {
		return fmt.Errorf("at least one instance is required")
	}

	// 验证每个实例的配置
	for i, instance := range config.Instances {
		if instance.RepositoryName == "" {
			return fmt.Errorf("instances[%d].repository_name is required", i)
		}
		if instance.BranchName == "" {
			return fmt.Errorf("instances[%d].branch_name is required", i)
		}
		if instance.WorkflowFileName == "" {
			return fmt.Errorf("instances[%d].workflow_file_name is required", i)
		}
		if instance.ComposeFilePath == "" {
			return fmt.Errorf("instances[%d].compose_file_path is required", i)
		}
		if instance.ComposeProjectDir == "" {
			return fmt.Errorf("instances[%d].compose_project_dir is required", i)
		}
	}

	return nil
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

// matchesConditions 检查 payload 是否匹配实例配置的条件
func matchesConditions(payload *WorkflowRunPayload, instance *InstanceConfig) bool {
	// 检查仓库名称
	if payload.Repository.FullName != instance.RepositoryName {
		log.Printf("Repository mismatch: got %s, expected %s", payload.Repository.FullName, instance.RepositoryName)
		return false
	}

	// 检查分支名称
	if payload.WorkflowRun.HeadBranch != instance.BranchName {
		log.Printf("Branch mismatch: got %s, expected %s", payload.WorkflowRun.HeadBranch, instance.BranchName)
		return false
	}

	// 检查工作流文件名
	if !strings.HasSuffix(payload.WorkflowRun.Path, instance.WorkflowFileName) {
		log.Printf("Workflow file mismatch: got %s, expected suffix %s", payload.WorkflowRun.Path, instance.WorkflowFileName)
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

// findMatchingInstance 查找匹配的部署实例
func findMatchingInstance(payload *WorkflowRunPayload, config *Config) *InstanceConfig {
	for i := range config.Instances {
		if matchesConditions(payload, &config.Instances[i]) {
			return &config.Instances[i]
		}
	}
	return nil
}

// executeDockerCompose 执行 Docker Compose 部署命令
func executeDockerCompose(instance *InstanceConfig) error {
	log.Printf("Starting Docker Compose deployment for %s...", instance.RepositoryName)

	// 1. 拉取最新镜像
	log.Println("Pulling latest images...")
	if err := runDockerComposeCommand(instance, "pull"); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	// 2. 停止并删除现有容器
	log.Println("Stopping and removing existing containers...")
	if err := runDockerComposeCommand(instance, "down"); err != nil {
		return fmt.Errorf("failed to stop containers: %w", err)
	}

	// 3. 重新创建并启动容器
	log.Println("Starting containers...")
	if err := runDockerComposeCommand(instance, "up", "-d"); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}

	log.Println("Docker Compose deployment completed successfully")
	return nil
}

// runDockerComposeCommand 运行 docker compose 命令
func runDockerComposeCommand(instance *InstanceConfig, args ...string) error {
	// 构建命令参数
	cmdArgs := []string{"compose", "-f", instance.ComposeFilePath}
	cmdArgs = append(cmdArgs, args...)

	// 创建命令
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Dir = instance.ComposeProjectDir
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
		if !verifySignature(config.Server.WebhookSecret, signature, body) {
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

		// 查找匹配的部署实例
		instance := findMatchingInstance(&payload, config)
		if instance == nil {
			log.Println("No matching instance found, skipping deployment")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "No matching instance found")
			return
		}

		// 执行部署
		log.Printf("Found matching instance for %s/%s, starting deployment...",
			instance.RepositoryName, instance.BranchName)
		if err := executeDockerCompose(instance); err != nil {
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

	log.Printf("Starting webhook server on port %s", config.Server.Port)
	log.Printf("Loaded %d deployment instance(s):", len(config.Instances))
	for i, instance := range config.Instances {
		log.Printf("  [%d] Repository: %s, Branch: %s, Workflow: %s",
			i+1,
			instance.RepositoryName,
			instance.BranchName,
			instance.WorkflowFileName,
		)
	}

	// 设置路由
	http.HandleFunc("/webhook", webhookHandler(config))
	http.HandleFunc("/health", healthHandler)

	// 启动服务器
	addr := ":" + config.Server.Port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
