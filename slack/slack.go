package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab.com/superman-tech/lib/env"
	"gitlab.com/superman-tech/lib/httprq"
)

const (
	defaultTimeout = 10
)

func NewFromConfig(cfg *SlackConfig) Slacker {
	return New(cfg.ServiceName, cfg.WebhookURL, WithTimeout(cfg.Timeout),
		WithDisable(cfg.Disable))
}

// New create new instance of slack client, serviceName will be appended to each message
// a valid slack webhookURL must be provided.
func New(serviceName, webhookURL string, opts ...Option) Slacker {
	option := slackOption{
		disable: false,
	}
	for _, o := range opts {
		o(&option)
	}
	if option.disable {
		return &noopSlack{}
	}
	if option.timeout < 0 {
		option.timeout = defaultTimeout
	}

	return newSlackImpl(serviceName, webhookURL, &option)
}

func newSlackImpl(serviceName, webhookURL string, o *slackOption) *slackImpl {
	host, _ := os.Hostname()
	env := env.Get()
	startTime := time.Now().Format(time.RFC3339)
	slack := &slackImpl{
		header: fmt.Sprintf("*[%s]* *Service*: `%s` *Host*: `%s` *Start Time*: `%s`",
			env, serviceName, host, startTime),
		webhookURL: webhookURL,
		timeout:    o.timeout,
	}
	return slack
}

// PostToSlack send message to preconfigure slack webhook, if needed post option can be used
// To provide more information for the message. please be aware of the context parameter
// Context must not be canceled or reach deadline, otherwise request may failed.
func (s *slackImpl) PostToSlack(ctx context.Context, message string, opts ...PostOption) error {
	payload := payload{
		Text:    message,
		Context: s.header,
		Fields:  map[string]interface{}{},
	}

	for _, opt := range opts {
		opt(&payload)
	}

	payloadBody := buildPayloadBody(&payload)

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payloadBody); err != nil {
		return fmt.Errorf("failed to encode payload body: %w", err)
	}

	res := httprq.Post(s.webhookURL).WithContext(ctx).WithTimeout(s.timeout).
		WithBody(buf).Execute()

	if res.Error != nil {
		return fmt.Errorf("failed to send webhook request: %w", res.Error)
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("failed to send slack message, status: %d body: %s", res.StatusCode,
		res.Body.String())
}

func buildPayloadBody(payload *payload) map[string]interface{} {
	blocks := headerBlock(payload.Text, payload.Context)

	if len(payload.Fields) > 0 {
		blocks = append(blocks, fieldsBlock(payload.Fields)...)
	}

	if len(payload.StackTrace) > 0 {
		blocks = append(blocks, stacktraceBlock(payload.StackTrace)...)
	}

	return map[string]interface{}{
		"text":   payload.Text,
		"blocks": blocks,
	}
}

func headerBlock(title, context string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "context",
			"elements": []map[string]interface{}{
				{
					"type": "mrkdwn",
					"text": context,
				},
			},
		},
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": title,
			},
		},
		{
			"type": "divider",
		},
	}
}

func fieldsBlock(fields map[string]interface{}) []map[string]interface{} {
	title := "*Fields*"
	content := fmt.Sprintf("```%s```", fieldsToString(fields))
	return sectionBlock(title, content)
}

func fieldsToString(fields map[string]interface{}) string {
	arr := make([]string, 0)
	for field, value := range fields {
		arr = append(arr, fmt.Sprintf("%s: %v", field, value))
	}

	return strings.Join(arr, "\n")
}

func stacktraceBlock(stacktrace string) []map[string]interface{} {
	title := "*Stack Trace* :bangbang:"
	content := fmt.Sprintf("```%s```", stacktrace)
	return sectionBlock(title, content)
}

func sectionBlock(title, content string) []map[string]interface{} {
	text := fmt.Sprintf("%s\n%s", title, content)
	return []map[string]interface{}{
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": text,
			},
		},
		{
			"type": "divider",
		},
	}
}
