package notifications

import (
	"context"
	"fmt"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/msgpack"
	"strconv"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/internal/email"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/queue"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
)

type NotificationType string

const (
	SlackNotificationType NotificationType = "slack"
	EmailNotificationType NotificationType = "email"
)

type Notification struct {
	// Defines the type of notification either slack or email.
	NotificationType NotificationType `json:"notification_type,omitempty"`

	// Email or Slack notification
	Payload interface{} `json:"payload,omitempty"`
}

type SlackNotification struct {
	WebhookURL string `json:"webhook_url,omitempty"`

	Text string `json:"text,omitempty"`
}

// NOTIFICATIONS

func SendEndpointNotification(_ context.Context,
	endpoint *datastore.Endpoint,
	project *datastore.Project,
	status datastore.EndpointStatus,
	q queue.Queuer,
	failure bool,
	failureMsg string,
	responseBody string,
	statusCode int,
) error {
	var ns []*Notification

	if !util.IsStringEmpty(endpoint.SupportEmail) {
		ns = append(ns, &Notification{NotificationType: EmailNotificationType})
	}

	if !util.IsStringEmpty(endpoint.SlackWebhookURL) {
		ns = append(ns, &Notification{NotificationType: SlackNotificationType})
	}

	for _, v := range ns {
		switch v.NotificationType {
		case EmailNotificationType:
			v.Payload = email.Message{
				Email:        endpoint.SupportEmail,
				Subject:      "Endpoint Status Update",
				TemplateName: email.TemplateEndpointUpdate,
				Params: map[string]string{
					"logo_url":        project.LogoURL,
					"target_url":      endpoint.TargetURL,
					"failure_msg":     failureMsg,
					"response_body":   responseBody,
					"status_code":     strconv.Itoa(statusCode),
					"endpoint_status": string(status),
				},
			}

		case SlackNotificationType:
			payload := SlackNotification{
				WebhookURL: endpoint.SlackWebhookURL,
			}

			var text string
			if failure {
				text = fmt.Sprintf("failed to send event delivery to endpoint url (%s) after retry limit was hit, endpoint response body (%s) and status code was %d, reason for failure is \"%s\", endpoint status is now %s", endpoint.TargetURL, responseBody, statusCode, failureMsg, status)
			} else {
				text = fmt.Sprintf("endpoint url (%s) which was formerly dectivated has now been reactivated, endpoint status is now %s", endpoint.TargetURL, status)
			}

			payload.Text = text
			v.Payload = payload
		default:
			log.Error("Invalid notification type")
			continue
		}

		buf, err := msgpack.EncodeMsgPack(v)
		if err != nil {
			log.WithError(err).Errorf("Failed to marshal %v notification payload", v.NotificationType)
			continue
		}

		job := &queue.Job{
			Payload: buf,
			Delay:   0,
		}

		err = q.Write(convoy.NotificationProcessor, convoy.DefaultQueue, job)
		if err != nil {
			log.WithError(err).Error("Failed to write new notification to the queue")
		}
	}

	return nil
}
