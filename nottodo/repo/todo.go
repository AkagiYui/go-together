package repo

// Todo 结构体定义
type Todo struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Completed   bool   `json:"completed"`
    CreatedAt   string `json:"created_at"`
}

// 基于 PostgreSQL 的 CRUD 实现

func GetTodos() ([]Todo, int) {
    rows, err := db.Query(`SELECT id, title, description, completed, created_at FROM todos ORDER BY id`)
    if err != nil {
        return []Todo{}, 0
    }
    list, err := scanTodoRows(rows)
    if err != nil {
        return []Todo{}, 0
    }
    return list, len(list)
}

func GetTodoByID(id int) (Todo, bool) {
    row := db.QueryRow(`SELECT id, title, description, completed, created_at FROM todos WHERE id = $1`, id)
    t, err := scanTodoRow(row)
    if err != nil {
        return Todo{}, false
    }
    return t, true
}

func UpdateTodo(todo Todo) bool {
    res, err := db.Exec(`UPDATE todos SET title = $1, description = $2, completed = $3 WHERE id = $4`,
        todo.Title, todo.Description, todo.Completed, todo.ID)
    if err != nil {
        return false
    }
    affected, _ := res.RowsAffected()
    return affected > 0
}

func DeleteTodo(id int) bool {
    res, err := db.Exec(`DELETE FROM todos WHERE id = $1`, id)
    if err != nil {
        return false
    }
    affected, _ := res.RowsAffected()
    return affected > 0
}

func CreateTodo(todo Todo) (Todo, bool) {
    row := db.QueryRow(`INSERT INTO todos (title, description, completed) VALUES ($1, $2, $3) RETURNING id, created_at`,
        todo.Title, todo.Description, todo.Completed)
    var t Todo
    var err error
    t, err = scanTodoRow(row)
    if err != nil {
        return Todo{}, false
    }
    // 将返回的字段合并到结果
    t.Title = todo.Title
    t.Description = todo.Description
    t.Completed = todo.Completed
    return t, true
}
