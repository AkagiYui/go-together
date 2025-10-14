package repo

import "time"

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

func GetTodos() ([]Todo, int) {
	return todos, len(todos)
}

func GetTodoByID(id int) (Todo, bool) {
	for _, todo := range todos {
		if todo.ID == id {
			return todo, true
		}
	}
	return Todo{}, false
}

func UpdateTodo(todo Todo) bool {
	for i, t := range todos {
		if t.ID == todo.ID {
			todos[i] = todo
			return true
		}
	}
	return false
}

func DeleteTodo(id int) bool {
	for i, todo := range todos {
		if todo.ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			return true
		}
	}
	return false
}

func CreateTodo(todo Todo) bool {
	todo.ID = nextID
	todo.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	nextID++
	todos = append(todos, todo)
	return true
}
