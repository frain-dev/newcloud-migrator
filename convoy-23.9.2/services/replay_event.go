package services

import (
	"context"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/msgpack"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/queue"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/worker/task"
)

type ReplayEventService struct {
	EndpointRepo datastore.EndpointRepository
	Queue        queue.Queuer

	Event *datastore.Event
}

func (e *ReplayEventService) Run(ctx context.Context) error {
	taskName := convoy.CreateEventProcessor

	createEvent := task.CreateEvent{
		Event: *e.Event,
	}

	eventByte, err := msgpack.EncodeMsgPack(createEvent)
	if err != nil {
		return &ServiceError{ErrMsg: err.Error()}
	}

	job := &queue.Job{
		ID:      e.Event.UID,
		Payload: eventByte,
		Delay:   0,
	}

	err = e.Queue.Write(taskName, convoy.CreateEventQueue, job)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("replay_event: failed to write event to the queue")
		return &ServiceError{ErrMsg: "failed to write event to queue", Err: err}
	}

	return nil
}
