package gorep

import "github.com/jmoiron/sqlx"

type Database interface {
	sqlx.Queryer
	sqlx.Execer
}
