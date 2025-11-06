package main

// 构建信息变量，通过 -ldflags 在编译时注入
var (
	// GitCommitHash Git 提交哈希值
	GitCommitHash = "unknown"
	// BuildTime 构建时间
	BuildTime = "unknown"
	// CommitTime Git 提交时间
	CommitTime = "unknown"
)

// BuildInfo 构建信息结构体
type BuildInfo struct {
	GitCommitHash string `json:"git_commit_hash"`
	BuildTime     string `json:"build_time"`
	CommitTime    string `json:"commit_time"`
}

// GetBuildInfo 获取构建信息
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		GitCommitHash: GitCommitHash,
		BuildTime:     BuildTime,
		CommitTime:    CommitTime,
	}
}

