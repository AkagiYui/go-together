module github.com/akagiyui/go-together/nottodo

go 1.24.5

require (
    github.com/akagiyui/go-together/common v0.0.0
    github.com/akagiyui/go-together/rest v0.0.0
    github.com/jackc/pgx/v5 v5.6.0 // indirect
    golang.org/x/crypto v0.27.0
)

replace github.com/akagiyui/go-together/common => ../common

replace github.com/akagiyui/go-together/rest => ../rest
