package worker

import (
	"context"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/apm"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/queue"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/worker/task"
	"github.com/hibiken/asynq"
)

type Consumer struct {
	queue queue.Queuer
	mux   *asynq.ServeMux
	srv   *asynq.Server
	log   log.StdLogger
}

func NewConsumer(q queue.Queuer, lo log.StdLogger) *Consumer {
	var opts asynq.RedisConnOpt

	if len(q.Options().RedisAddress) == 1 {
		opts = q.Options().RedisClient
	} else {
		opts = asynq.RedisClusterClientOpt{
			Addrs: q.Options().RedisAddress,
		}
	}

	srv := asynq.NewServer(
		opts,
		asynq.Config{
			Concurrency: convoy.Concurrency,
			BaseContext: func() context.Context {
				return log.NewContext(context.Background(), lo, nil)
			},
			Queues: q.Options().Names,
			IsFailure: func(err error) bool {
				if _, ok := err.(*task.RateLimitError); ok {
					return false
				}
				return true
			},
			RetryDelayFunc: task.GetRetryDelay,
			Logger:         lo,
		},
	)

	mux := asynq.NewServeMux()

	return &Consumer{
		queue: q,
		log:   lo,
		mux:   mux,
		srv:   srv,
	}
}

func (c *Consumer) Start() {
	if err := c.srv.Start(c.mux); err != nil {
		c.log.WithError(err).Fatal("error starting worker")
	}
}

func (c *Consumer) RegisterHandlers(taskName convoy.TaskName, handlerFn func(context.Context, *asynq.Task) error) {
	c.mux.HandleFunc(string(taskName), c.loggingMiddleware(asynq.HandlerFunc(handlerFn)).ProcessTask)
}

func (c *Consumer) Stop() {
	c.srv.Stop()
	c.srv.Shutdown()
}

func (c *Consumer) loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		txn, innerCtx := apm.StartTransaction(ctx, t.Type())
		defer txn.End()

		err := h.ProcessTask(innerCtx, t)
		if err != nil {
			c.log.WithError(err).WithField("job", t.Type()).Error("job failed")
			return err
		}
		return nil
	})
}