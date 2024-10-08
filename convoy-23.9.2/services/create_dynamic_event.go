package services

import (
	"context"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/msgpack"
	"net/http"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/models"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/queue"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
	"github.com/google/uuid"
)

type CreateDynamicEventService struct {
	Queue queue.Queuer

	DynamicEvent *models.DynamicEvent
	Project      *datastore.Project
}

func (e *CreateDynamicEventService) Run(ctx context.Context) error {
	if e.Project == nil {
		return &ServiceError{ErrMsg: "an error occurred while creating event - invalid project"}
	}

	e.DynamicEvent.Event.ProjectID = e.Project.UID

	taskName := convoy.CreateDynamicEventProcessor

	eventByte, err := msgpack.EncodeMsgPack(e.DynamicEvent)
	if err != nil {
		return util.NewServiceError(http.StatusBadRequest, err)
	}

	job := &queue.Job{
		ID:      uuid.NewString(),
		Payload: eventByte,
		Delay:   0,
	}

	err = e.Queue.Write(taskName, convoy.CreateEventQueue, job)
	if err != nil {
		log.FromContext(ctx).Errorf("Error occurred sending new dynamic event to the queue %s", err)
		return &ServiceError{ErrMsg: "failed to create dynamic event"}
	}

	return nil
}
