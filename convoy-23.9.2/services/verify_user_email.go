package services

import (
	"context"
	"time"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
)

type VerifyEmailService struct {
	UserRepo datastore.UserRepository
	Token    string
}

func (u *VerifyEmailService) Run(ctx context.Context) error {
	user, err := u.UserRepo.FindUserByEmailVerificationToken(ctx, u.Token)
	if err != nil {
		if err == datastore.ErrUserNotFound {
			return &ServiceError{ErrMsg: "invalid password reset token"}
		}

		log.FromContext(ctx).WithError(err).Error("failed to find user by email verification token")
		return &ServiceError{ErrMsg: "failed to find user", Err: err}
	}

	if time.Now().After(user.EmailVerificationExpiresAt) {
		return &ServiceError{ErrMsg: "email verification token has expired"}
	}

	user.EmailVerified = true
	err = u.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("failed to update user")
		return &ServiceError{ErrMsg: "failed to update user", Err: err}
	}

	return nil
}
