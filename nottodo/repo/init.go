package repo

import (
	"context"

	"github.com/akagiyui/go-together/nottodo/config"
	"github.com/jackc/pgx/v5"
)

var (
	Db   *Queries
	conn *pgx.Conn
	Ctx  = context.Background()
)

func init() {
	var err error
	conn, err = pgx.Connect(Ctx, config.GlobalConfig.DSN)
	if err != nil {
		panic(err)
	}
	Db = New(conn)
}
