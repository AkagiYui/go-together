package repo

import (
	"context"

	todogen "github.com/akagiyui/go-together/nottodo/repo/todo"
)

// 兼容旧的仓储接口，内部转调 sqlc 生成的查询
// 注意：这里不再编写任何原生 SQL

type Todo = todogen.Todo

func GetTodos(ctx context.Context) ([]Todo, int, error) {
	list, err := TodoQueries.ListTodos(ctx)
	if err != nil {
		return nil, 0, err
	}
	return list, len(list), nil
}

func GetTodoByID(ctx context.Context, id int) (Todo, bool, error) {
	t, err := TodoQueries.GetTodo(ctx, id)
	if err != nil {
		return Todo{}, false, err
	}
	return t, true, nil
}

func UpdateTodo(ctx context.Context, todo Todo) (bool, error) {
	affected, err := TodoQueries.UpdateTodo(ctx, todo.ID, todo.Title, todo.Description, todo.Completed)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func DeleteTodo(ctx context.Context, id int) (bool, error) {
	affected, err := TodoQueries.DeleteTodo(ctx, id)
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func CreateTodo(ctx context.Context, todo Todo) (Todo, bool, error) {
	created, err := TodoQueries.CreateTodo(ctx, todo.Title, todo.Description, todo.Completed)
	if err != nil {
		return Todo{}, false, err
	}
	return created, true, nil
}
