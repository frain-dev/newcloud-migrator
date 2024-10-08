package main

import (
	"context"
	"fmt"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/postgres"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
)

func (m *Migrator) RunEndpointMigration() error {
	endpointRepo := postgres.NewEndpointRepo(m)

	for _, p := range m.projects {
		endpoints, err := m.loadProjectEndpoints(endpointRepo, p.UID, defaultPageable)
		if err != nil {
			return err
		}

		if len(endpoints) > 0 {
			err = m.SaveEndpoints(context.Background(), endpoints)
			if err != nil {
				return fmt.Errorf("failed to save endpoints: %v", err)
			}
		}

		for _, endpoint := range endpoints {
			m.endpointIDs[endpoint.UID] = struct{}{}
		}
	}

	return nil
}

const (
	saveEndpoints = `
	INSERT INTO convoy.endpoints (
		id, name, status, secrets, owner_id, url, description, http_timeout,
		rate_limit, rate_limit_duration, advanced_signatures, slack_webhook_url,
		support_email, app_id, project_id, authentication_type, authentication_type_api_key_header_name,
		authentication_type_api_key_header_value, created_at, updated_at, deleted_at
	)
	VALUES
	  (
		:id, :name, :status, :secrets, :owner_id, :url, :description, :http_timeout,
		:rate_limit, :rate_limit_duration, :advanced_signatures, :slack_webhook_url,
		:support_email, :app_id, :project_id, :authentication_type, :authentication_type_api_key_header_name,
		:authentication_type_api_key_header_value, :created_at, :updated_at, :deleted_at
	  )
	`
)

func (m *Migrator) SaveEndpoints(ctx context.Context, endpoints []datastore.Endpoint) error {
	values := make([]map[string]interface{}, 0, len(endpoints))

	for i := range endpoints {
		endpoint := &endpoints[i]

		ac := endpoint.GetAuthConfig()

		values = append(values, map[string]interface{}{
			"id":                  endpoint.UID,
			"name":                endpoint.Title,
			"status":              endpoint.Status,
			"secrets":             endpoint.Secrets,
			"owner_id":            endpoint.OwnerID,
			"url":                 endpoint.TargetURL,
			"description":         endpoint.Description,
			"http_timeout":        10,
			"rate_limit":          0,
			"rate_limit_duration": 0,
			"advanced_signatures": true,
			"slack_webhook_url":   endpoint.SlackWebhookURL,
			"support_email":       endpoint.SupportEmail,
			"app_id":              endpoint.AppID,
			"project_id":          endpoint.ProjectID,
			"authentication_type": ac.Type,
			"authentication_type_api_key_header_name":  ac.ApiKey.HeaderName,
			"authentication_type_api_key_header_value": ac.ApiKey.HeaderValue,
			"created_at": endpoint.CreatedAt,
			"updated_at": endpoint.UpdatedAt,
			"deleted_at": endpoint.DeletedAt,
		})
	}

	_, err := m.newDB.NamedExecContext(ctx, saveEndpoints, values)
	return err
}
