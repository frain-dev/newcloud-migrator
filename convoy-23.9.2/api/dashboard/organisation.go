package dashboard

import (
	"net/http"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/models"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/postgres"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/services"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
	"github.com/go-chi/render"

	m "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/middleware"
)

func (a *DashboardHandler) GetOrganisation(w http.ResponseWriter, r *http.Request) {
	org, err := a.retrieveOrganisation(r)
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	if err = a.A.Authz.Authorize(r.Context(), "organisation.manage", org); err != nil {
		_ = render.Render(w, r, util.NewErrorResponse("Unauthorized", http.StatusForbidden))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Organisation fetched successfully", org, http.StatusOK))
}

func (a *DashboardHandler) GetOrganisationsPaged(w http.ResponseWriter, r *http.Request) { // TODO: change to GetUserOrganisationsPaged
	pageable := m.GetPageableFromContext(r.Context())
	user, err := a.retrieveUser(r)
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	organisations, paginationData, err := postgres.NewOrgMemberRepo(a.A.DB).LoadUserOrganisationsPaged(r.Context(), user.UID, pageable)
	if err != nil {
		log.FromContext(r.Context()).WithError(err).Error("failed to fetch user organisations")
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Organisations fetched successfully",
		pagedResponse{Content: &organisations, Pagination: &paginationData}, http.StatusOK))
}

func (a *DashboardHandler) CreateOrganisation(w http.ResponseWriter, r *http.Request) {
	var newOrg models.Organisation
	err := util.ReadJSON(r, &newOrg)
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	user, err := a.retrieveUser(r)
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	co := services.CreateOrganisationService{
		OrgRepo:       postgres.NewOrgRepo(a.A.DB),
		OrgMemberRepo: postgres.NewOrgMemberRepo(a.A.DB),
		NewOrg:        &newOrg,
		User:          user,
	}

	organisation, err := co.Run(r.Context())
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Organisation created successfully", organisation, http.StatusCreated))
}

func (a *DashboardHandler) UpdateOrganisation(w http.ResponseWriter, r *http.Request) {
	var orgUpdate models.Organisation
	err := util.ReadJSON(r, &orgUpdate)
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	org, err := a.retrieveOrganisation(r)
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	if err = a.A.Authz.Authorize(r.Context(), "organisation.manage", org); err != nil {
		_ = render.Render(w, r, util.NewErrorResponse("Unauthorized", http.StatusForbidden))
		return
	}

	us := services.UpdateOrganisationService{
		OrgRepo:       postgres.NewOrgRepo(a.A.DB),
		OrgMemberRepo: postgres.NewOrgMemberRepo(a.A.DB),
		Org:           org,
		Update:        &orgUpdate,
	}

	org, err = us.Run(r.Context())
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Organisation updated successfully", org, http.StatusAccepted))
}

func (a *DashboardHandler) DeleteOrganisation(w http.ResponseWriter, r *http.Request) {
	org, err := a.retrieveOrganisation(r)
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	if err = a.A.Authz.Authorize(r.Context(), "organisation.manage", org); err != nil {
		_ = render.Render(w, r, util.NewErrorResponse("Unauthorized", http.StatusForbidden))
		return
	}

	err = postgres.NewOrgRepo(a.A.DB).DeleteOrganisation(r.Context(), org.UID)
	if err != nil {
		log.FromContext(r.Context()).WithError(err).Error("failed to delete organisation")
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Organisation deleted successfully", nil, http.StatusOK))
}
