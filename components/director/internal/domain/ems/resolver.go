package ems

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Quality int

const (
	Acknowledge    Quality = 1
	NonAcknowledge Quality = 0
)

type Webhook struct {
	Name             string     `json:"name"`
	Address          string     `json:"address"`
	QualityOfService Quality    `json:"qos"`
	PushConfig       PushConfig `json:"pushConfig"`
}

type PushConfig struct {
	Type            string          `json:"type"`
	Endpoint        string          `json:"endpoint"`
	ExemptHandshake bool            `json:"exemptHandshake"`
	SecuritySchema  *SecuritySchema `json:"securitySchema,omitempty"`
}

type SecuritySchema struct {
	Type         string `json:"type"`
	User         string `json:"user"`
	Password     string `json:"password"`
	GrantType    string `json:"grantType"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	TokenURL     string `json:"tokenUrl"`
}

type Service interface {
	CreateQueue(ctx context.Context, name string) (string, error)
	CreateTopicSubscription(ctx context.Context, queueName string, topic string) error
	CreateWebhook(ctx context.Context, queueName string, webhook *Webhook) error
}

type Resolver struct {
	emsSvc Service
}

func NewResolver(emsSvc Service) *Resolver {
	return &Resolver{
		emsSvc: emsSvc,
	}
}

func (r *Resolver) RequestEventSubscription(ctx context.Context, subject string, webhook *graphql.EventWebhook) (*graphql.EventSubscription, error) {
	queueName, err := r.emsSvc.CreateQueue(ctx, subject)
	if err != nil {
		return nil, errors.Wrap(err, "while creating queue")
	}

	if err := r.emsSvc.CreateTopicSubscription(ctx, subject, subject+"-brokers"); err != nil {
		return nil, errors.Wrap(err, "while creating topic subscription")
	}

	if webhook != nil {
		wh := &Webhook{
			Name:             webhook.Name,
			Address:          subject,
			QualityOfService: Acknowledge,
			PushConfig: PushConfig{
				Endpoint:        webhook.URL,
				ExemptHandshake: false,
				Type:            "webhook",
				SecuritySchema:  &SecuritySchema{},
			},
		}

		switch webhook.AuthType {
		case graphql.AuthTypeBasic:
			wh.PushConfig.SecuritySchema.Type = "basicAuth"
			wh.PushConfig.SecuritySchema.User = *webhook.Username
			wh.PushConfig.SecuritySchema.Password = *webhook.Password
		case graphql.AuthTypeOAuth2:
			wh.PushConfig.SecuritySchema.Type = "oauth2"
			wh.PushConfig.SecuritySchema.TokenURL = *webhook.TokenURL
			wh.PushConfig.SecuritySchema.ClientID = *webhook.ClientID
			wh.PushConfig.SecuritySchema.ClientSecret = *webhook.ClientSecret
			wh.PushConfig.SecuritySchema.GrantType = "client_credentials"
		default:
			wh.PushConfig.SecuritySchema = nil
		}

		if err := r.emsSvc.CreateWebhook(ctx, subject, wh); err != nil {
			return nil, apperrors.NewInvalidDataError("cannot create webhook %s", err.Error())
		}
	}
	result := &graphql.EventSubscription{
		Address: queueName,
	}

	return result, nil
}
