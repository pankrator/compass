package ems

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	MessagingURL   string
	ManagementURL  string
	MessagingPath  string
	ManagementPath string
	TokenURL       string
	ClientID       string
	ClientSecret   string
	Namespace      string
}

type EMSService struct {
	httpClient *http.Client
	Config     *Config
}

func NewService(ctx context.Context, httpClient *http.Client, config *Config) *EMSService {
	return &EMSService{
		httpClient: getHTTPClient(ctx, httpClient, config),
		Config:     config,
	}
}

type AccessType string

const (
	ExclusiveAccess    AccessType = "EXCLUSIVE"
	NonExclusiveAccess AccessType = "NON_EXCLUSIVE"
)

type QueueConfig struct {
	AccessType                     AccessType `json:"accessType"`
	MaxDeliveredUnackedMsgsPerFlow int64      `json:"maxDeliveredUnackedMsgsPerFlow"`
	MaxMessageSizeInBytes          int64      `json:"maxMessageSizeInBytes"`
	MaxQueueMessageCount           int64      `json:"maxQueueMessageCount"`
	MaxQueueSizeInBytes            int64      `json:"maxQueueSizeInBytes"`
	RespectTtl                     bool       `json:"respectTtl"`
}

func getDefaultQueueConfig() *QueueConfig {
	return &QueueConfig{
		AccessType:                     NonExclusiveAccess,
		MaxDeliveredUnackedMsgsPerFlow: 10000,
		MaxMessageSizeInBytes:          10000000,
		MaxQueueMessageCount:           1000,
		MaxQueueSizeInBytes:            1572864000,
		RespectTtl:                     false,
	}
}

func (es *EMSService) CreateQueue(ctx context.Context, name string) (string, error) {
	queueConfig := getDefaultQueueConfig()
	fullQueueName := es.Config.Namespace + "/" + name
	bodyBytes, err := json.Marshal(queueConfig)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling queue config")
	}
	reader := bytes.NewReader(bodyBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s%s/%s", es.Config.ManagementURL, es.Config.ManagementPath, url.PathEscape(fullQueueName)), reader)
	if err != nil {
		return "", errors.Wrap(err, "while creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "while requesting ems")
	}
	if resp.StatusCode == http.StatusCreated {
		log.Infof("Queue with name %s was successfully created", name)
		return fullQueueName, nil
	} else if resp.StatusCode == http.StatusOK {
		log.Infof("Queue with name %s was successfully updated", name)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", errors.New("could not create queue " + string(body))
	}

	return fullQueueName, nil
}

func (es *EMSService) CreateTopicSubscription(ctx context.Context, queueName string, topic string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s%s/%s/subscriptions/%s", es.Config.ManagementURL, es.Config.ManagementPath, url.PathEscape(es.Config.Namespace+"/"+queueName), url.PathEscape(topic)), nil)
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "while requesting ems")
	}
	if resp.StatusCode == http.StatusCreated {
		log.Infof("Subscription for queue %s was successfully created", queueName)
		return nil
	} else if resp.StatusCode == http.StatusOK {
		log.Infof("Subscription for queue %s was successfully updated", queueName)
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New("could not create topic subscription " + string(body))
	}

	return nil
}

func (es *EMSService) CreateWebhook(ctx context.Context, queueName string, webhook *Webhook) error {
	webhook.Address = "queue:" + es.Config.Namespace + "/" + queueName
	bodyBytes, err := json.Marshal(webhook)
	if err != nil {
		return errors.Wrap(err, "while marshaling queue config")
	}

	// TODO: Check if such webhook already exists and do not allow creation - otherwise there might be 2 webhooks and only one will be called which may be problematic
	reader := bytes.NewReader(bodyBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s%s", es.Config.MessagingURL, es.Config.MessagingPath), reader)
	if err != nil {
		return errors.Wrap(err, "while creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := es.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "while requesting ems")
	}
	if resp.StatusCode == http.StatusAccepted {
		log.Infof("Webhook %s was successfully created", webhook.Name)
		return nil
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New("could not create webhook " + resp.Status + " " + string(body))
	}

}

func getHTTPClient(ctx context.Context, httpClient *http.Client, config *Config) *http.Client {
	cc := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     config.TokenURL,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)
	return cc.Client(ctx)
}
