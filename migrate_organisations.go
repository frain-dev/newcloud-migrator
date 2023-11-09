package main

import (
	"context"
	"fmt"
	ncache "github.com/frain-dev/convoy/cache/noop"
	"github.com/frain-dev/convoy/database/postgres"
	"github.com/frain-dev/convoy/datastore"
	"github.com/frain-dev/convoy/util"
)

func (m *Migrator) RunOrgMigration() error {
	orgs, err := m.loadOrganisations(pagedResponse{
		Data: data{
			Pagination: datastore.PaginationData{PerPage: 1000},
		},
	})
	if err != nil {
		return err
	}

	// filter orgs owned by the user
	userOrgs := []datastore.Organisation{}
	for _, org := range orgs {
		if org.OwnerID == m.user.UID {
			userOrgs = append(userOrgs, org)
		}
	}

	m.userOrgs = userOrgs
	if len(userOrgs) == 0 {
		return fmt.Errorf("user does not own any orgs")
	}

	orgMemberRepo := postgres.NewOrgMemberRepo(m, ncache.NewNoopCache())
	userIDs := map[string]struct{}{}
	members := []*datastore.OrganisationMember{}
	for _, org := range userOrgs {
		mm, err := m.loadOrgMembers(orgMemberRepo, org.UID, defaultPageable)
		if err != nil {
			return fmt.Errorf("failed to load org %s members: %v", org.UID, err)
		}

		for _, mr := range mm {
			userIDs[mr.UserID] = struct{}{}
		}

		fmt.Println("mm", mm)
		members = append(members, mm...)
	}

	users, err := m.loadUsers(userIDs)
	if err != nil {
		return fmt.Errorf("failed to load org member users: %v", err)
	}

	err = m.SaveUsers(context.Background(), users)
	if err != nil {
		return fmt.Errorf("failed to save org memeber users: %v", err)
	}

	err = m.SaveOrganisations(context.Background(), userOrgs)
	if err != nil {
		return fmt.Errorf("failed to save orgs: %v", err)
	}

	err = m.SaveOrganisationMembers(context.Background(), members)
	if err != nil {
		return fmt.Errorf("failed to save orgs: %v", err)
	}

	return nil
}

const (
	saveOrganizations = `
	INSERT INTO convoy.organisations (id, name, owner_id, custom_domain, assigned_domain, created_at, updated_at, deleted_at)
	VALUES (
	    :id, :name, :owner_id, :custom_domain, :assigned_domain, :created_at, :updated_at, :deleted_at
	)
	`
)

func (m *Migrator) SaveOrganisations(ctx context.Context, orgs []datastore.Organisation) error {
	values := make([]map[string]interface{}, 0, len(orgs))

	for _, org := range orgs {
		values = append(values, map[string]interface{}{
			"id":              org.UID,
			"name":            org.Name,
			"owner_id":        org.OwnerID,
			"custom_domain":   org.CustomDomain,
			"assigned_domain": org.AssignedDomain,
			"created_at":      org.CreatedAt,
			"updated_at":      org.UpdatedAt,
			"deleted_at":      org.DeletedAt,
		})
	}

	_, err := m.newDB.NamedExecContext(ctx, saveOrganizations, values)
	if err != nil {
		return fmt.Errorf("failed to save orgs: %v", err)
	}
	return nil
}

const (
	saveOrgMembers = `
	INSERT INTO convoy.organisation_members (id, organisation_id, user_id, role_type, role_project, role_endpoint, created_at, updated_at, deleted_at)
	VALUES (
	    :id, :organisation_id, :user_id, :role_type, :role_project,
	    :role_endpoint, :created_at, :updated_at, :deleted_at
	)
	`
)

func (o *Migrator) SaveOrganisationMembers(ctx context.Context, members []*datastore.OrganisationMember) error {
	values := make([]map[string]interface{}, 0, len(members))

	for _, member := range members {
		var endpointID *string
		var projectID *string
		if !util.IsStringEmpty(member.Role.Endpoint) {
			endpointID = &member.Role.Endpoint
		}

		if !util.IsStringEmpty(member.Role.Project) {
			projectID = &member.Role.Project
		}

		values = append(values, map[string]interface{}{
			"id":              member.UID,
			"organisation_id": member.OrganisationID,
			"user_id":         member.UserID,
			"role_type":       member.Role.Type,
			"role_project":    projectID,
			"role_endpoint":   endpointID,
			"created_at":      member.CreatedAt,
			"updated_at":      member.UpdatedAt,
			"deleted_at":      member.DeletedAt,
		})
	}

	_, err := o.newDB.NamedExecContext(ctx, saveOrgMembers, values)
	return err
}
