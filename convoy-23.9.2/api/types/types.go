package types

import (
	authz "github.com/Subomi/go-authz"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/cache"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/queue"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/tracer"
)

type ContextKey string

type APIOptions struct {
	DB     database.Database
	Queue  queue.Queuer
	Logger log.StdLogger
	Tracer tracer.Tracer
	Cache  cache.Cache
	Authz  *authz.Authz
}
