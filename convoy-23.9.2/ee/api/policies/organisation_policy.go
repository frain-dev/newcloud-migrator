package policies

import (
	"context"
	"errors"

	authz "github.com/Subomi/go-authz"
	basepolicy "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/policies"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/auth"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
)

type OrganisationPolicy struct {
	*authz.BasePolicy
	OrganisationMemberRepo datastore.OrganisationMemberRepository
}

func (op *OrganisationPolicy) Manage(ctx context.Context, res interface{}) error {
	authCtx := ctx.Value(basepolicy.AuthUserCtx).(*auth.AuthenticatedUser)

	org, ok := res.(*datastore.Organisation)
	if !ok {
		return errors.New("Wrong organisation type")
	}

	// Dashboard Access or Personal Access Token

	user, ok := authCtx.User.(*datastore.User)
	if !ok {
		return basepolicy.ErrNotAllowed
	}

	member, err := op.OrganisationMemberRepo.FetchOrganisationMemberByUserID(ctx, user.UID, org.UID)
	if err != nil {
		return basepolicy.ErrNotAllowed
	}

	if isAllowed := isSuperAdmin(member); !isAllowed {
		return basepolicy.ErrNotAllowed
	}

	return nil
}

func (op *OrganisationPolicy) GetName() string {
	return "organisation"
}
