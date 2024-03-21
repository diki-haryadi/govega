package slack

import (
	"context"
	"fmt"
)

type (
	Slacker interface {
		PostToSlack(ctx context.Context, message string, opts ...PostOption) error
	}

	SlackConfig struct {
		ServiceName string `json:"service_name" mapstructure:"service_name"`
		WebhookURL  string `json:"webhook_url" mapstructure:"webhook_url"`
		Timeout     int    `json:"timeout" mapstructure:"timeout"`
		Disable     bool   `json:"disable" mapstructure:"disable"`
	}

	slackImpl struct {
		timeout    int
		webhookURL string
		header     string
	}
	slackOption struct {
		timeout int
		disable bool
	}
	Option func(s *slackOption)

	payload struct {
		Text       string
		Context    string
		StackTrace string
		Fields     map[string]interface{}
	}
	PostOption func(p *payload)

	noopSlack struct{}
)

// MentionUser utility to add user mention tag to message
// Please refer to https://api.slack.com/reference/surfaces/formatting#mentioning-users
func MentionUser(id string) string {
	return fmt.Sprintf("<@%s>", id)
}

// MentionGroup utility to add group mention tag to message
// Please refer to https://api.slack.com/reference/surfaces/formatting#mentioning-groups
func MentionGroup(id string) string {
	return fmt.Sprintf("<!subteam^%s>", id)
}

// MentionHere utility to add here mention, to notify active user in channel
func MentionHere() string {
	return "<!here>"
}

// MentionHere utility to add channel mention, to notify all user in channel
func MentionChannel() string {
	return "<!channel>"
}

// WithTimeout define slack client timeout in seconds, Default to 10 seconds
func WithTimeout(timeout int) Option {
	return func(s *slackOption) {
		s.timeout = timeout
	}
}

func WithDisable(disable bool) Option {
	return func(s *slackOption) {
		s.disable = true
	}
}

// WithStacktrace append stack trace to slack message
func WithStacktrace(stacktrace string) PostOption {
	return func(p *payload) {
		p.StackTrace = stacktrace
	}
}

// WithField append field and it's value to slack message
func WithField(name string, value interface{}) PostOption {
	return func(p *payload) {
		p.Fields[name] = value
	}
}

// WithFields append fields to slack message
func WithFields(fields map[string]interface{}) PostOption {
	return func(p *payload) {
		p.Fields = fields
	}
}
