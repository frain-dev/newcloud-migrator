package main

import (
	"context"
	"database/sql"
	ncache "github.com/frain-dev/convoy/cache/noop"
	"github.com/frain-dev/convoy/database/postgres"

	"github.com/frain-dev/convoy/datastore"
	"github.com/frain-dev/convoy/util"
)

func (m *Migrator) RunEventMigration() error {
	eventRepo := postgres.NewEventRepo(m, ncache.NewNoopCache())

	for _, p := range m.projects {
		err := m.loadEvents(eventRepo, p, defaultPageable)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	saveEvents = `
	INSERT INTO convoy.events (
	id, event_type, endpoints, project_id, source_id, headers,
	raw, data,created_at,updated_at, deleted_at, url_query_params,
    idempotency_key, is_duplicate_event
    )
	VALUES (
	    :id, :event_type, :endpoints, :project_id, :source_id,
	    :headers, :raw, :data, :created_at, :updated_at, :deleted_at, :url_query_params,
        :idempotency_key,:is_duplicate_event
	)
	`

	createEventEndpoints = `
	INSERT INTO convoy.events_endpoints (endpoint_id, event_id) VALUES (:endpoint_id, :event_id)
	`
)

func (e *Migrator) SaveEvents(ctx context.Context, events []datastore.Event) error {
	ev := make([]map[string]interface{}, 0, len(events))
	evEndpoints := make([]postgres.EventEndpoint, 0, len(events)*2)

	dedupe := map[string]int{}

	for i := range events {
		event := &events[i]

		if _, ok := e.eventIDs[event.UID]; ok { //if previously saved, ignore
			continue
		}

		switch dedupe[event.UID] {
		case 0:
			dedupe[event.UID] = 1
		case 1:
			continue
		}

		var sourceID *string

		for _, endpointID := range event.Endpoints {
			if _, ok := e.endpointIDs[endpointID]; !ok {
				continue
			}
		}

		if !util.IsStringEmpty(event.SourceID) {
			if _, ok := e.sourceIDs[event.SourceID]; !ok {
				continue
			}

			sourceID = &event.SourceID
		}

		ev = append(ev, map[string]interface{}{
			"id":                 event.UID,
			"event_type":         event.EventType,
			"endpoints":          event.Endpoints,
			"project_id":         event.ProjectID,
			"source_id":          sourceID,
			"headers":            event.Headers,
			"raw":                event.Raw,
			"data":               event.Data,
			"created_at":         event.CreatedAt,
			"updated_at":         event.UpdatedAt,
			"deleted_at":         event.DeletedAt,
			"url_query_params":   event.URLQueryParams,
			"idempotency_key":    event.IdempotencyKey,
			"is_duplicate_event": event.IsDuplicateEvent,
		})

		if len(event.Endpoints) > 0 {
			for _, endpointID := range event.Endpoints {
				evEndpoints = append(evEndpoints, postgres.EventEndpoint{EventID: event.UID, EndpointID: endpointID})
			}
		}
	}

	tx, err := e.newDB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer rollbackTx(tx)

	if len(ev) > 0 {
		_, err = tx.NamedExecContext(ctx, saveEvents, ev)
		if err != nil {
			return err
		}

		if len(evEndpoints) > 0 {
			_, err = tx.NamedExecContext(ctx, createEventEndpoints, evEndpoints)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
