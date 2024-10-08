package task

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/email"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/pkg/smtp"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/msgpack"
	"github.com/hibiken/asynq"
)

var ErrInvalidEmailPayload = errors.New("invalid email payload")

func ProcessEmails(sc smtp.SmtpClient) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var message email.Message

		err := msgpack.DecodeMsgPack(t.Payload(), &message)
		if err != nil {
			err := json.Unmarshal(t.Payload(), &message)
			if err != nil {
				return ErrInvalidEmailPayload
			}
		}

		newEmail := email.NewEmail(sc)

		if err := newEmail.Build(string(message.TemplateName), message.Params); err != nil {
			return err
		}

		if err := newEmail.Send(message.Email, message.Subject); err != nil {
			return err
		}

		return nil
	}
}
