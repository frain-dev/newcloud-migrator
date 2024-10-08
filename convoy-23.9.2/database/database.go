package database

import (
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/hooks"
	"github.com/jmoiron/sqlx"
)

type Database interface {
	GetDB() *sqlx.DB
	GetHook() *hooks.Hook
	Close() error
}
