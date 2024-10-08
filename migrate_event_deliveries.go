package main

import (
	"context"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/postgres"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
)

func (m *Migrator) RunEventDeliveriesMigration() error {
	eventDeliveryRepo := postgres.NewEventDeliveryRepo(m)

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

func (m *Migrator) SaveEventDeliveries(ctx context.Context, deliveries []datastore.EventDelivery) error {
	values := make([]map[string]interface{}, 0, len(deliveries))
	dedupe := map[string]int{}

	for i := range deliveries {
		delivery := &deliveries[i]

		if _, ok := m.deliveryIDs[delivery.UID]; ok { // if previously saved, ignore
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
			if _, ok := m.endpointIDs[delivery.EndpointID]; !ok {
				continue
			}

			endpointID = &delivery.EndpointID
		}

		if _, ok := m.eventIDs[delivery.EventID]; !ok {
			continue
		}

		if _, ok := m.subIDs[delivery.SubscriptionID]; !ok {
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
		_, err := m.newDB.NamedExecContext(ctx, saveEventDeliveries, values)
		return err
	}
	return nil
}
