package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/resolute-sh/resolute/core"
)

// BlockType identifies the type of Slack Block Kit block.
type BlockType string

const (
	BlockTypeSection BlockType = "section"
	BlockTypeHeader  BlockType = "header"
	BlockTypeDivider BlockType = "divider"
)

// TextType identifies the type of text in a Slack block.
type TextType string

const (
	TextTypeMrkdwn    TextType = "mrkdwn"
	TextTypePlainText TextType = "plain_text"
)

// Text represents a text object in Slack Block Kit.
type Text struct {
	Type TextType `json:"type"`
	Text string   `json:"text"`
}

// Block represents a Slack Block Kit block.
type Block struct {
	Type BlockType `json:"type"`
	Text *Text     `json:"text,omitempty"`
}

// SendMessageInput is the input for SendMessageActivity.
type SendMessageInput struct {
	WebhookURL string
	Channel    string
	Text       string
	Blocks     []Block
}

// SendMessageOutput is the output of SendMessageActivity.
type SendMessageOutput struct {
	Sent bool
}

type webhookPayload struct {
	Channel string  `json:"channel,omitempty"`
	Text    string  `json:"text,omitempty"`
	Blocks  []Block `json:"blocks,omitempty"`
}

// httpPoster abstracts HTTP POST for testing.
type httpPoster interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// SendMessageActivity sends a message to Slack via webhook.
func SendMessageActivity(ctx context.Context, input SendMessageInput) (SendMessageOutput, error) {
	return sendMessageWithClient(ctx, http.DefaultClient, input)
}

func sendMessageWithClient(_ context.Context, client httpPoster, input SendMessageInput) (SendMessageOutput, error) {
	if input.Text == "" && len(input.Blocks) == 0 {
		return SendMessageOutput{Sent: false}, nil
	}

	payload := webhookPayload{
		Channel: input.Channel,
		Text:    input.Text,
		Blocks:  input.Blocks,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SendMessageOutput{}, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := client.Post(input.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return SendMessageOutput{}, fmt.Errorf("post to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return SendMessageOutput{}, fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return SendMessageOutput{Sent: true}, nil
}

// SendMessage creates a node for the SendMessageActivity.
func SendMessage(input SendMessageInput) *core.Node[SendMessageInput, SendMessageOutput] {
	return core.NewNode("slack.SendMessage", SendMessageActivity, input)
}
