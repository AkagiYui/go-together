package repo

import "time"

func GetTodos() ([]Todo, int64, error) {
	todos, err := Db.ListTodos(Ctx)
	if err != nil {
		return nil, 0, err
	}
	total, err := Db.CountTodos(Ctx)
	return todos, total, err
}

func GetTodoByID(id int64) (Todo, error) {
	todo, err := Db.GetTodo(Ctx, id)
	if err != nil {
		return Todo{}, err
	}
	return todo, nil
}

func UpdateTodo(todo Todo) bool {
	updateParam := UpdateTodoParams{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
	}

	if err := Db.UpdateTodo(Ctx, updateParam); err != nil {
		return false
	}
	return true
}

func DeleteTodo(id int64) error {
	err := Db.DeleteTodo(Ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func CreateTodo(todo Todo) (Todo, error) {
	todo.CreatedAt.Time = time.Now()
	todo.CreatedAt.Valid = true
	return Db.CreateTodo(Ctx, CreateTodoParams{
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
	})
}
