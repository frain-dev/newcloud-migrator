package dashboard

import (
	"net/http"
	"time"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/auth/realm/jwt"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/config"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/database/postgres"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/services"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/models"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/middleware"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
	"github.com/go-chi/render"
)

type AuthorizedLogin struct {
	Username   string    `json:"username,omitempty"`
	Token      string    `json:"token"`
	ExpiryTime time.Time `json:"expiry_time"`
}

func (a *DashboardHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var newUser models.LoginUser
	if err := util.ReadJSON(r, &newUser); err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	config, err := config.Get()
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	lu := services.LoginUserService{
		UserRepo: postgres.NewUserRepo(a.A.DB),
		Cache:    a.A.Cache,
		JWT:      jwt.NewJwt(&config.Auth.Jwt, a.A.Cache),
		Data:     &newUser,
	}

	user, token, err := lu.Run(r.Context())
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusUnauthorized))
		return
	}

	u := &models.LoginUserResponse{
		User:  user,
		Token: models.Token{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken},
	}

	_ = render.Render(w, r, util.NewServerResponse("Login successful", u, http.StatusOK))
}

func (a *DashboardHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshToken models.Token
	if err := util.ReadJSON(r, &refreshToken); err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	if err := refreshToken.Validate(); err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	config, err := config.Get()
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	rf := services.RefreshTokenService{
		UserRepo: postgres.NewUserRepo(a.A.DB),
		JWT:      jwt.NewJwt(&config.Auth.Jwt, a.A.Cache),
		Data:     &refreshToken,
	}

	token, err := rf.Run(r.Context())
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusUnauthorized))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Token refresh successful", token, http.StatusOK))
}

func (a *DashboardHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	auth, err := middleware.GetAuthFromRequest(r)
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusUnauthorized))
		return
	}

	config, err := config.Get()
	if err != nil {
		_ = render.Render(w, r, util.NewErrorResponse(err.Error(), http.StatusBadRequest))
		return
	}

	lg := services.LogoutUserService{
		UserRepo: postgres.NewUserRepo(a.A.DB),
		JWT:      jwt.NewJwt(&config.Auth.Jwt, a.A.Cache),
		Token:    auth.Token,
	}

	err = lg.Run(r.Context())
	if err != nil {
		_ = render.Render(w, r, util.NewServiceErrResponse(err))
		return
	}

	_ = render.Render(w, r, util.NewServerResponse("Logout successful", nil, http.StatusOK))
}