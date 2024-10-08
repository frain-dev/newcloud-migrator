package policies

import (
	"context"
	"errors"

	authz "github.com/Subomi/go-authz"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/auth"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
)

type ProjectPolicy struct {
	*authz.BasePolicy
	OrganisationRepo       datastore.OrganisationRepository
	OrganisationMemberRepo datastore.OrganisationMemberRepository
}

func (pp *ProjectPolicy) Manage(ctx context.Context, res interface{}) error {
	authCtx := ctx.Value(AuthUserCtx).(*auth.AuthenticatedUser)

	project, ok := res.(*datastore.Project)
	if !ok {
		return errors.New("Wrong project type")
	}

	org, err := pp.OrganisationRepo.FetchOrganisationByID(ctx, project.OrganisationID)
	if err != nil {
		return ErrNotAllowed
	}

	apiKey, ok := authCtx.APIKey.(*datastore.APIKey)
	if ok {
		// Personal Access Tokens
		if apiKey.Type == datastore.PersonalKey {
			_, err := pp.OrganisationMemberRepo.FetchOrganisationMemberByUserID(ctx, apiKey.UserID, org.UID)
			if err != nil {
				return ErrNotAllowed
			}

			return nil
		}

		// API Key
		if apiKey.Role.Project != project.UID {
			return ErrNotAllowed
		}

		return nil
	}

	// JWT Access.
	orgPolicy := OrganisationPolicy{
		OrganisationMemberRepo: pp.OrganisationMemberRepo,
	}
	return orgPolicy.Manage(ctx, org)
}

func (pp *ProjectPolicy) GetName() string {
	return "project"
}
