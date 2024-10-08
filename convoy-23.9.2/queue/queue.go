package queue

import (
	"time"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/rdb"
)

type Queuer interface {
	Write(convoy.TaskName, convoy.QueueName, *Job) error
	Options() QueueOptions
}

type Job struct {
	ID      string        `json:"id"`
	Payload []byte        `json:"payload"`
	Delay   time.Duration `json:"delay"`
}

type QueueOptions struct {
	Names             map[string]int
	Type              string
	RedisClient       *rdb.Redis
	RedisAddress      []string
	PrometheusAddress string
}
