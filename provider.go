// Package slack provides Slack integration activities for resolute workflows.
package slack

import (
	"github.com/resolute-sh/resolute/core"
	"go.temporal.io/sdk/worker"
)

const (
	ProviderName    = "resolute-slack"
	ProviderVersion = "0.1.0"
)

// Provider returns the Slack provider for registration.
func Provider() core.Provider {
	return core.NewProvider(ProviderName, ProviderVersion).
		AddActivity("slack.SendMessage", SendMessageActivity)
}

// RegisterActivities registers all Slack activities with a Temporal worker.
func RegisterActivities(w worker.Worker) {
	core.RegisterProviderActivities(w, Provider())
}
