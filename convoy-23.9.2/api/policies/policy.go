package policies

import (
	"errors"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/types"
)

const AuthUserCtx types.ContextKey = "authUser"

// ErrNotAllowed is returned when request is not permitted.
var ErrNotAllowed = errors.New("unauthorized to process request")
