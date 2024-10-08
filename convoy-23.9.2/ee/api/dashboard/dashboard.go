package dashboard

import (
	base "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/dashboard"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/types"
)

type DashboardHandler struct {
	*base.DashboardHandler
	Opts *types.APIOptions
}
