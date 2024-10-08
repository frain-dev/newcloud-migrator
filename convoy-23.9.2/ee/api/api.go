package api

import (
	"net/http"

	authz "github.com/Subomi/go-authz"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api"
	base "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/dashboard"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/types"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/postgres"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/ee/api/dashboard"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/ee/api/policies"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

type EHandler struct {
	*api.ApplicationHandler
	opts *types.APIOptions
}

func NewEHandler(opts *types.APIOptions) (*EHandler, error) {
	az, err := authz.NewAuthz(&authz.AuthzOpts{
		AuthCtxKey: authz.AuthCtxType(middleware.AuthUserCtx),
	})

	if err != nil {
		return &EHandler{}, err
	}

	opts.Authz = az
	eeh := &EHandler{
		opts:               opts,
		ApplicationHandler: &api.ApplicationHandler{A: opts},
	}

	return eeh, nil
}

func (eh *EHandler) BuildRoutes() http.Handler {
	// register community routes
	router := eh.ApplicationHandler.BuildRoutes()

	// apply overrides
	eh.RegisterEnterpriseDashboardHandler(router)

	return router
}

func (eh *EHandler) RegisterEnterpriseDashboardHandler(r *chi.Mux) {
	_ = &dashboard.DashboardHandler{
		DashboardHandler: base.NewDashboardHandler(eh.opts),
		Opts:             eh.opts,
	}
}

func (eh *EHandler) RegisterPolicy() error {
	var err error

	err = eh.opts.Authz.RegisterPolicy(func() authz.Policy {
		po := &policies.ProjectPolicy{
			BasePolicy:             authz.NewBasePolicy(),
			OrganisationRepo:       postgres.NewOrgRepo(eh.opts.DB),
			OrganisationMemberRepo: postgres.NewOrgMemberRepo(eh.opts.DB),
		}

		po.SetRule("manage", authz.RuleFunc(po.Manage))

		return po
	}())

	if err != nil {
		return err
	}

	err = eh.opts.Authz.RegisterPolicy(func() authz.Policy {
		po := &policies.OrganisationPolicy{
			BasePolicy:             authz.NewBasePolicy(),
			OrganisationMemberRepo: postgres.NewOrgMemberRepo(eh.opts.DB),
		}

		po.SetRule("manage", authz.RuleFunc(po.Manage))

		return po
	}())

	return err
}