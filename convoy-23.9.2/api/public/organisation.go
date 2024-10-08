package public

import (
	"net/http"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/postgres"
	m "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/middleware"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
	"github.com/go-chi/render"
)

func (a *PublicHandler) GetOrganisationsPaged(w http.ResponseWriter, r *http.Request) { // TODO: change to GetUserOrganisationsPaged
	pageable := m.GetPageableFromContext(r.Context())
	user, err := a.retrieveUser(r)
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	organisations, paginationData, err := postgres.NewOrgMemberRepo(a.A.DB).LoadUserOrganisationsPaged(r.Context(), user.UID, pageable)
	if err != nil {
		log.FromContext(r.Context()).WithError(err).Error("failed to fetch user organisations")
		_ = render.Render(w, r, util.NewErrorResponse("failed to fetch user organisations", http.StatusBadRequest))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Organisations fetched successfully",
		pagedResponse{Content: &organisations, Pagination: &paginationData}, http.StatusOK))
}
