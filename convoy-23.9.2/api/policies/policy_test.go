package policies

import (
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/auth"
	"github.com/stretchr/testify/require"
)

type basetest struct {
	name          string
	authCtx       *auth.AuthenticatedUser
	assertion     require.ErrorAssertionFunc
	expectedError error
}
