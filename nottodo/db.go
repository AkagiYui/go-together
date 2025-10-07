package main

// Todo 结构体定义
type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
}

// 内存数据库 - 用于存储Todo项目
var todos []Todo
var nextID = 1

func init() {
	// 初始化一些示例数据
	todos = []Todo{
		{
			ID:          1,
			Title:       "学习Go语言",
			Description: "完成Go语言基础教程",
			Completed:   false,
			CreatedAt:   "2024-01-01 10:00:00",
		},
		{
			ID:          2,
			Title:       "学习Gin框架",
			Description: "掌握Gin框架的基本用法",
			Completed:   true,
			CreatedAt:   "2024-01-02 11:00:00",
		},
	}
	nextID = 3
}
