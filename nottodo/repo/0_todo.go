package repo

import "time"

// Todo 待办事项表
type Todo struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                                                       // ID
	Title       string    `gorm:"column:title;type:varchar(255);not null" json:"title"`                                                               // 标题
	Description *string   `gorm:"column:description;type:varchar(255)" json:"description"`                                                            // 描述（可空）
	Completed   bool      `gorm:"column:completed;not null;index:idx_todos_completed" json:"completed"`                                               // 是否完成
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp;index:idx_todos_created_at" json:"createdAt"` // 创建时间（非空）
}

// TableName 指定表名
func (Todo) TableName() string {
	return "todos"
}

// GetTodos 获取所有待办事项
func GetTodos() ([]Todo, int64, error) {
	var todos []Todo
	var total int64

	// 查询总数
	if err := DB.Model(&Todo{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表，按 ID 排序
	if err := DB.Order("id").Find(&todos).Error; err != nil {
		return nil, 0, err
	}

	return todos, total, nil
}

// GetTodoByID 根据ID获取待办事项
func GetTodoByID(id int64) (Todo, error) {
	var todo Todo
	result := DB.First(&todo, id)
	return todo, result.Error
}

// UpdateTodo 更新待办事项
func UpdateTodo(todo Todo) error {
	// 使用 Updates 更新多个字段
	return DB.Model(&Todo{}).Where("id = ?", todo.ID).Updates(map[string]any{
		"title":       todo.Title,
		"description": todo.Description,
		"completed":   todo.Completed,
	}).Error
}

// DeleteTodo 删除待办事项
func DeleteTodo(id int64) error {
	return DB.Delete(&Todo{}, id).Error
}

// CreateTodo 创建待办事项
func CreateTodo(todo Todo) (Todo, error) {
	result := DB.Create(&todo)
	return todo, result.Error
}
