package policies

import (
	"context"
	"errors"

	authz "github.com/Subomi/go-authz"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/auth"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
)

type OrganisationPolicy struct {
	*authz.BasePolicy
	OrganisationMemberRepo datastore.OrganisationMemberRepository
}

func (op *OrganisationPolicy) Manage(ctx context.Context, res interface{}) error {
	authCtx := ctx.Value(AuthUserCtx).(*auth.AuthenticatedUser)

	user, ok := authCtx.User.(*datastore.User)
	if !ok {
		return ErrNotAllowed
	}

	org, ok := res.(*datastore.Organisation)
	if !ok {
		return errors.New("Wrong organisation type")
	}

	member, err := op.OrganisationMemberRepo.FetchOrganisationMemberByUserID(ctx, user.UID, org.UID)
	if err != nil {
		return ErrNotAllowed
	}

	if member.Role.Type != auth.RoleSuperUser {
		return ErrNotAllowed
	}

	return nil
}

func (op *OrganisationPolicy) GetName() string {
	return "organisation"
}
