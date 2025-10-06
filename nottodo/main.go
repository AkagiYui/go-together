package main

import (
	"time"

	"github.com/akagiyui/go-together/rest"
)

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

type CreateTodoRequest struct {
	Todo
}

func (r *CreateTodoRequest) Handle(ctx *rest.Context) any {
	// 验证必填字段
	if r.Title == "" {
		return map[string]string{
			"error": "Title is required",
		}
	}

	// 创建新的 Todo
	newTodo := Todo{
		ID:          nextID,
		Title:       r.Title,
		Description: r.Description,
		Completed:   r.Completed,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	nextID++
	todos = append(todos, newTodo)

	return map[string]interface{}{
		"message": "Todo created successfully",
		"data":    newTodo,
	}
}

type GetTodosRequest struct{}

func (r *GetTodosRequest) Handle(ctx *rest.Context) any {
	return map[string]interface{}{
		"data":  todos,
		"count": len(todos),
	}
}

type GetTodoByIDRequest struct {
	ID int `path:"id"`
}

func (r *GetTodoByIDRequest) Handle(ctx *rest.Context) any {
	for _, todo := range todos {
		if todo.ID == r.ID {
			return todo
		}
	}
	return map[string]string{
		"error": "Todo not found",
	}
}

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

func main() {
	s := rest.NewServer()

	s.GETFunc("/healtz", func(ctx *rest.Context) any {
		return "Hello, World!"
	})
	s.POST("/todos", &CreateTodoRequest{})
	s.GET("/todos", &GetTodosRequest{})
	s.GET("/todos/{id}", &GetTodoByIDRequest{})

	println("🚀 Server starting on http://localhost:8080")
	println("📚 API Documentation:")
	println("  POST   /todos        - 创建Todo")
	println("  GET    /todos        - 获取所有Todo")
	println("  GET    /todos/{id}   - 获取指定ID的Todo")

	if err := s.Run(":8080"); err != nil {
		panic(err)
	}
}
