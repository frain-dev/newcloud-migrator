package main

import (
	"context"
	"database/sql"
	"fmt"
	ncache "github.com/frain-dev/convoy/cache/noop"
	"github.com/frain-dev/convoy/database/postgres"

	"github.com/frain-dev/convoy/datastore"
	"github.com/frain-dev/convoy/util"
)

func (m *Migrator) RunSourceMigration() error {
	sourceRepo := postgres.NewSourceRepo(m, ncache.NewNoopCache())
	for _, p := range m.projects {
		sources, err := m.loadProjectSources(sourceRepo, p.UID, defaultPageable)
		if err != nil {
			return err
		}

		if len(sources) > 0 {
			err = m.SaveSources(context.Background(), sources)
			if err != nil {
				return fmt.Errorf("failed to save sources: %v", err)
			}
		}
	}

	return nil
}

const (
	saveSources = `
    INSERT INTO convoy.sources
        (id, source_verifier_id, name,type,mask_id,provider,
        is_disabled,forward_headers,project_id, pub_sub, created_at, updated_at,
        custom_response_body, custom_response_content_type, idempotency_keys
         )
    VALUES (
        :id, :source_verifier_id, :name, :type, :mask_id, :provider,
        :is_disabled, :forward_headers, :project_id, :pub_sub, :created_at, :updated_at,
        :custom_response_body, :custom_response_content_type, :idempotency_keys
    )
    `

	saveSourceVerifiers = `
    INSERT INTO convoy.source_verifiers (
        id,type,basic_username,basic_password,
        api_key_header_name,api_key_header_value,
        hmac_hash,hmac_header,hmac_secret,hmac_encoding
    )
    VALUES (
        :id, :type, :basic_username, :basic_password,
        :api_key_header_name, :api_key_header_value,
        :hmac_hash, :hmac_header, :hmac_secret, :hmac_encoding
    )
    `
)

func (s *Migrator) SaveSources(ctx context.Context, sources []datastore.Source) error {
	sourceValues := make([]map[string]interface{}, 0, len(sources))
	sourceVerifierValues := make([]map[string]interface{}, 0, len(sources))

	for _, source := range sources {
		var (
			sourceVerifierID *string
			hmac             datastore.HMac
			basic            datastore.BasicAuth
			apiKey           datastore.ApiKey
		)

		switch source.Verifier.Type {
		case datastore.APIKeyVerifier:
			apiKey = *source.Verifier.ApiKey
		case datastore.BasicAuthVerifier:
			basic = *source.Verifier.BasicAuth
		case datastore.HMacVerifier:
			hmac = *source.Verifier.HMac
		}

		if !util.IsStringEmpty(string(source.Verifier.Type)) {
			sourceVerifierID = &source.VerifierID

			sourceVerifierValues = append(sourceVerifierValues, map[string]interface{}{
				"id":                   sourceVerifierID,
				"type":                 source.Verifier.Type,
				"basic_username":       basic.UserName,
				"basic_password":       basic.Password,
				"api_key_header_name":  apiKey.HeaderName,
				"api_key_header_value": apiKey.HeaderValue,
				"hmac_hash":            hmac.Hash,
				"hmac_header":          hmac.Header,
				"hmac_secret":          hmac.Secret,
				"hmac_encoding":        hmac.Encoding,
			})
		}

		sourceValues = append(sourceValues, map[string]interface{}{
			"id":                           source.UID,
			"source_verifier_id":           sourceVerifierID,
			"name":                         source.Name,
			"type":                         source.Type,
			"mask_id":                      source.MaskID,
			"provider":                     source.Provider,
			"is_disabled":                  source.IsDisabled,
			"forward_headers":              source.ForwardHeaders,
			"project_id":                   source.ProjectID,
			"pub_sub":                      source.PubSub,
			"created_at":                   source.CreatedAt,
			"updated_at":                   source.UpdatedAt,
			"custom_response_body":         source.CustomResponse.Body,
			"custom_response_content_type": source.CustomResponse.ContentType,
			"idempotency_keys":             source.IdempotencyKeys,
		})
	}

	tx, err := s.newDB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

	if len(sourceVerifierValues) > 0 {
		_, err = tx.NamedExecContext(ctx, saveSourceVerifiers, sourceVerifierValues)
		if err != nil {
			return err
		}
	}

	_, err = tx.NamedExecContext(ctx, saveSources, sourceValues)
	if err != nil {
		return err
	}

	return tx.Commit()
}
