package repo

import "time"

// GetTodos 获取所有待办事项
func GetTodos() ([]Todo, int64, error) {
	todos, err := Db.ListTodos(Ctx)
	if err != nil {
		return nil, 0, err
	}
	total, err := Db.CountTodos(Ctx)
	return todos, total, err
}

// GetTodoByID 根据ID获取待办事项
func GetTodoByID(id int64) (Todo, error) {
	todo, err := Db.GetTodo(Ctx, id)
	if err != nil {
		return Todo{}, err
	}
	return todo, nil
}

// UpdateTodo 更新待办事项
func UpdateTodo(todo Todo) error {
	updateParam := UpdateTodoParams{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
	}

	return Db.UpdateTodo(Ctx, updateParam)
}

// DeleteTodo 删除待办事项
func DeleteTodo(id int64) error {
	err := Db.DeleteTodo(Ctx, id)
	if err != nil {
		return err
	}
	return nil
}

// CreateTodo 创建待办事项
func CreateTodo(todo Todo) (Todo, error) {
	todo.CreatedAt.Time = time.Now()
	todo.CreatedAt.Valid = true
	return Db.CreateTodo(Ctx, CreateTodoParams{
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
	})
}
