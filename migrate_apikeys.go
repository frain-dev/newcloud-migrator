package main

import (
	"context"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/auth"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/postgres"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
)

func (m *Migrator) RunAPIKeyMigration() error {
	apiKeyRepo := postgres.NewAPIKeyRepo(m)

	// migrate project api keys
	for _, p := range m.projects {
		err := m.loadAPIKeys(apiKeyRepo, p.UID, "", defaultPageable)
		if err != nil {
			return err
		}
	}

	// migrate user api keys
	err := m.loadAPIKeys(apiKeyRepo, "", m.user.UID, defaultPageable)
	if err != nil {
		return err
	}

	return nil
}

const (
	saveAPIKeys = `
    INSERT INTO convoy.api_keys (id,name,key_type,mask_id,role_type,role_project,role_endpoint,hash,salt,user_id,expires_at,created_at,updated_at, deleted_at)
    VALUES (
        :id, :name, :key_type, :mask_id, :role_type, :role_project,
        :role_endpoint, :hash, :salt, :user_id, :expires_at,
        :created_at, :updated_at, :deleted_at
    )
    `
)

func (m *Migrator) SaveAPIKeys(ctx context.Context, keys []datastore.APIKey) error {
	values := make([]map[string]interface{}, 0, len(keys))

	for i := range keys {
		key := &keys[i]
		var (
			userID     *string
			endpointID *string
			projectID  *string
			roleType   *auth.RoleType
		)

		if !util.IsStringEmpty(key.UserID) {
			userID = &key.UserID
		}

		if !util.IsStringEmpty(key.Role.Endpoint) {
			endpointID = &key.Role.Endpoint
		}

		if !util.IsStringEmpty(key.Role.Project) {
			projectID = &key.Role.Project
		}

		if !util.IsStringEmpty(string(key.Role.Type)) {
			roleType = &key.Role.Type
		}

		values = append(values, map[string]interface{}{
			"id":            key.UID,
			"name":          key.Name,
			"key_type":      key.Type,
			"mask_id":       key.MaskID,
			"role_type":     roleType,
			"role_project":  projectID,
			"role_endpoint": endpointID,
			"hash":          key.Hash,
			"salt":          key.Salt,
			"user_id":       userID,
			"expires_at":    key.ExpiresAt,
			"created_at":    key.CreatedAt,
			"updated_at":    key.UpdatedAt,
			"deleted_at":    key.DeletedAt,
		})
	}

	if len(values) > 0 {
		_, err := m.newDB.NamedExecContext(ctx, saveAPIKeys, values)
		return err
	}
	return nil
}
