module github.com/akagiyui/go-together/nottodo

go 1.24.5

require (
	github.com/akagiyui/go-together/common v0.0.0
	github.com/akagiyui/go-together/rest v0.0.0
	github.com/jackc/pgx/v5 v5.7.6
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
)

replace github.com/akagiyui/go-together/common => ../common

replace github.com/akagiyui/go-together/rest => ../rest
