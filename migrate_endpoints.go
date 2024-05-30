package main

import (
	"context"
	"fmt"
	ncache "github.com/frain-dev/convoy/cache/noop"
	"github.com/frain-dev/convoy/database/postgres"
	"github.com/frain-dev/convoy/datastore"
	"github.com/frain-dev/convoy/util"
	"strings"
	"time"
)

func (m *Migrator) RunEndpointMigration() error {
	endpointRepo := postgres.NewEndpointRepo(m, ncache.NewNoopCache())

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

		httpTimeout := int64(5)
		rateLimitDuration := int64(30)

		if !util.IsStringEmpty(endpoint.HttpTimeout) {
			v := endpoint.HttpTimeout
			if !strings.Contains(endpoint.HttpTimeout, "s") {
				v += "s"
			}

			d, err := time.ParseDuration(v)
			if err != nil {
				return fmt.Errorf("failed to parse endpoint HttpTimeout: %v", err)
			}
			httpTimeout = int64(d.Seconds())
		}

		if !util.IsStringEmpty(endpoint.RateLimitDuration) {
			d, err := time.ParseDuration(endpoint.RateLimitDuration)
			if err != nil {
				return fmt.Errorf("failed to parse endpoint HttpTimeout: %v", err)
			}

			rateLimitDuration = int64(d.Seconds())
		}

		values = append(values, map[string]interface{}{
			"id":                  endpoint.UID,
			"name":                endpoint.Title,
			"status":              endpoint.Status,
			"secrets":             endpoint.Secrets,
			"owner_id":            endpoint.OwnerID,
			"url":                 endpoint.TargetURL,
			"description":         endpoint.Description,
			"http_timeout":        httpTimeout,
			"rate_limit":          endpoint.RateLimit,
			"rate_limit_duration": rateLimitDuration,
			"advanced_signatures": endpoint.AdvancedSignatures,
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

const c = `
./nc run --old-base-url="https://dashboard.getconvoy.io" --pat="CO.5IdjnVLBm1BGtk1T.31vBixXVD4PzJGYgiRKqkhSIgrCF3OeSZHx2VyXsm7h4ZGwSktih77KK8cjKyI6E" --old-pg-dsn="postgresql://reader-writer:readWrite@convoy-prod-db.cqovqpuj1mkv.us-east-1.rds.amazonaws.com/convoy" --new-pg-dsn="postgres://dedicateddbadmin:qKQ71exN5irh0w5FrPrsflc97lDFmHm9@prod-dedicated-db-proxy-01.proxy-c0el2iadta4e.eu-west-1.rds.amazonaws.com:5432/dbkxbau9zvh3kh?sslmode=require&connect_timeout=30" --migrate-events

`
