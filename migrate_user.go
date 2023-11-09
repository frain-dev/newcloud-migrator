package main

import (
	"context"
	"github.com/frain-dev/convoy/datastore"
)

func (m *Migrator) RunUserMigration() error {
	user, err := m.loadUser()
	if err != nil {
		return err
	}
	m.user = user

	return nil
}

const (
	saveUsers = `
    INSERT INTO convoy.users (
		id,first_name,last_name,email,password,
        email_verified,reset_password_token, email_verification_token,
        reset_password_expires_at,email_verification_expires_at, created_at, updated_at, deleted_at)
    VALUES (
        :id, :first_name, :last_name, :email, :password,
        :email_verified, :reset_password_token, :email_verification_token,
        :reset_password_expires_at, :email_verification_expires_at, :created_at, :updated_at, :deleted_at
    )
    `
)

func (m *Migrator) SaveUsers(ctx context.Context, users []*datastore.User) error {
	values := make([]map[string]interface{}, 0, len(users))

	for _, user := range users {
		values = append(values, map[string]interface{}{
			"id":                            user.UID,
			"first_name":                    user.FirstName,
			"last_name":                     user.LastName,
			"email":                         user.Email,
			"password":                      user.Password,
			"email_verified":                user.EmailVerified,
			"reset_password_token":          user.ResetPasswordToken,
			"email_verification_token":      user.EmailVerificationToken,
			"reset_password_expires_at":     user.ResetPasswordExpiresAt,
			"email_verification_expires_at": user.EmailVerificationExpiresAt,
			"created_at":                    user.CreatedAt,
			"updated_at":                    user.UpdatedAt,
			"deleted_at":                    user.DeletedAt,
		})
	}

	_, err := m.newDB.NamedExecContext(ctx, saveUsers, values)
	return err
}
