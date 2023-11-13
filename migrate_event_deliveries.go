package main

import (
	"context"
	ncache "github.com/frain-dev/convoy/cache/noop"
	"github.com/frain-dev/convoy/database/postgres"

	"github.com/frain-dev/convoy/datastore"
	"github.com/frain-dev/convoy/util"
)

func (m *Migrator) RunEventDeliveriesMigration() error {
	eventDeliveryRepo := postgres.NewEventDeliveryRepo(m, ncache.NewNoopCache())

	for _, p := range m.projects {
		err := m.loadEventDeliveries(eventDeliveryRepo, p, defaultPageable)
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	saveEventDeliveries = `
    INSERT INTO convoy.event_deliveries (
          id, project_id, event_id, endpoint_id, subscription_id,
          headers, attempts, status, metadata, cli_metadata, description,
          created_at, updated_at, deleted_at
          )
    VALUES (
        :id, :project_id, :event_id, :endpoint_id,
        :subscription_id, :headers, :attempts, :status, :metadata,
        :cli_metadata, :description, :created_at, :updated_at, :deleted_at
    )
    `
)

func (e *Migrator) SaveEventDeliveries(ctx context.Context, deliveries []datastore.EventDelivery) error {
	values := make([]map[string]interface{}, 0, len(deliveries))
	dedupe := map[string]int{}

	for i := range deliveries {
		delivery := &deliveries[i]

		if _, ok := e.deliveryIDs[delivery.UID]; ok { //if previously saved, ignore
			continue
		}

		switch dedupe[delivery.UID] {
		case 0:
			dedupe[delivery.UID] = 1
		case 1:
			continue
		}

		var endpointID *string

		if !util.IsStringEmpty(delivery.EndpointID) {
			if _, ok := e.endpointIDs[delivery.EndpointID]; !ok {
				continue
			}

			endpointID = &delivery.EndpointID
		}

		if _, ok := e.eventIDs[delivery.EventID]; !ok {
			continue
		}

		if _, ok := e.subIDs[delivery.SubscriptionID]; !ok {
			continue
		}

		if !util.IsStringEmpty(delivery.DeviceID) {
			continue // ignore cli event deliveries
		}

		values = append(values, map[string]interface{}{
			"id":              delivery.UID,
			"project_id":      delivery.ProjectID,
			"event_id":        delivery.EventID,
			"endpoint_id":     endpointID,
			"subscription_id": delivery.SubscriptionID,
			"headers":         delivery.Headers,
			"attempts":        delivery.DeliveryAttempts,
			"status":          delivery.Status,
			"metadata":        delivery.Metadata,
			"cli_metadata":    delivery.CLIMetadata,
			"description":     delivery.Description,
			"created_at":      delivery.CreatedAt,
			"updated_at":      delivery.UpdatedAt,
			"deleted_at":      delivery.DeletedAt,
		})
	}

	if len(values) > 0 {
		_, err := e.newDB.NamedExecContext(ctx, saveEventDeliveries, values)
		return err
	}
	return nil
}
