package slack

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/resolute-sh/resolute/core"
)

const (
	bodyBlockCharLimit = 3000
	maxSlackBlocks     = 50
)

// NotifyReportInput is the input for NotifyReportActivity.
type NotifyReportInput struct {
	WebhookURL  string
	Header      string
	Label1      string
	Value1      string
	Label2      string
	Value2      string
	Body        string
	CostUSD     string
	Duration    string
	TurnsUsed   string
	Succeeded   string
	FailHeader  string
	FailMessage string
	LLMProvider string
	LLMModel    string
}

// NotifyReportOutput is the output of NotifyReportActivity.
type NotifyReportOutput struct {
	Notified bool
}

// NotifyReportActivity sends a Slack Block Kit report notification.
func NotifyReportActivity(ctx context.Context, input NotifyReportInput) (NotifyReportOutput, error) {
	if input.WebhookURL == "" {
		return NotifyReportOutput{Notified: false}, nil
	}

	var blocks []Block

	if input.Succeeded == "true" {
		blocks = append(blocks, Block{
			Type: BlockTypeHeader,
			Text: &Text{Type: TextTypePlainText, Text: input.Header},
		})

		meta := fmt.Sprintf("*%s:* %s  |  *%s:* %s", input.Label1, input.Value1, input.Label2, input.Value2)
		var metaParts []string
		if input.LLMProvider != "" || input.LLMModel != "" {
			provider := input.LLMProvider
			if provider == "" {
				provider = "anthropic"
			}
			metaParts = append(metaParts, fmt.Sprintf("LLM: %s / %s", provider, input.LLMModel))
		}
		if input.Duration != "" {
			metaParts = append(metaParts, fmt.Sprintf("Duration: %s", input.Duration))
		}
		if input.CostUSD != "" {
			metaParts = append(metaParts, fmt.Sprintf("Cost: $%s", input.CostUSD))
		}
		if input.TurnsUsed != "" {
			metaParts = append(metaParts, fmt.Sprintf("Turns: %s", input.TurnsUsed))
		}
		if len(metaParts) > 0 {
			meta += "\n" + strings.Join(metaParts, "  |  ")
		}
		blocks = append(blocks, Block{
			Type: BlockTypeSection,
			Text: &Text{Type: TextTypeMrkdwn, Text: meta},
		})

		blocks = append(blocks, Block{Type: BlockTypeDivider})

		body := markdownToMrkdwn(input.Body)
		blocks = append(blocks, formatBodyBlocks(body)...)
	} else {
		blocks = append(blocks,
			Block{
				Type: BlockTypeHeader,
				Text: &Text{Type: TextTypePlainText, Text: input.FailHeader},
			},
			Block{
				Type: BlockTypeSection,
				Text: &Text{Type: TextTypeMrkdwn, Text: input.FailMessage},
			},
		)
	}

	if len(blocks) > maxSlackBlocks {
		blocks = blocks[:maxSlackBlocks]
	}

	text := fmt.Sprintf("%s: %s (%s)", input.Header, input.Value1, input.Value2)

	result, err := SendMessageActivity(ctx, SendMessageInput{
		WebhookURL: input.WebhookURL,
		Text:       text,
		Blocks:     blocks,
	})
	if err != nil {
		return NotifyReportOutput{}, fmt.Errorf("send report notification: %w", err)
	}

	return NotifyReportOutput{Notified: result.Sent}, nil
}

var (
	reHeading  = regexp.MustCompile(`(?m)^#{1,3}\s+(.+)$`)
	reBold     = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reLink     = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reHRSplit  = regexp.MustCompile(`(?m)^---+\s*$`)
)

// markdownToMrkdwn converts common markdown syntax to Slack mrkdwn.
func markdownToMrkdwn(md string) string {
	s := reHeading.ReplaceAllString(md, "*$1*")
	s = reBold.ReplaceAllString(s, "*$1*")
	s = reLink.ReplaceAllString(s, "<$2|$1>")
	return strings.TrimSpace(s)
}

// formatBodyBlocks splits a converted mrkdwn body into Slack blocks,
// using --- as section dividers and respecting Slack's character limit.
func formatBodyBlocks(body string) []Block {
	if len(body) == 0 {
		return nil
	}

	sections := reHRSplit.Split(body, -1)
	var blocks []Block

	for i, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		if i > 0 {
			blocks = append(blocks, Block{Type: BlockTypeDivider})
		}

		blocks = append(blocks, splitBodyBlocks(section)...)
	}

	return blocks
}

// splitBodyBlocks splits text into section blocks within Slack's character limit.
func splitBodyBlocks(body string) []Block {
	if len(body) == 0 {
		return nil
	}

	var blocks []Block
	for len(body) > 0 {
		chunk := body
		if len(chunk) > bodyBlockCharLimit {
			chunk = body[:bodyBlockCharLimit]
		}
		body = body[len(chunk):]

		blocks = append(blocks, Block{
			Type: BlockTypeSection,
			Text: &Text{Type: TextTypeMrkdwn, Text: chunk},
		})
	}
	return blocks
}

// NotifyReport creates a node for NotifyReportActivity.
func NotifyReport(input NotifyReportInput) *core.Node[NotifyReportInput, NotifyReportOutput] {
	return core.NewNode("slack.NotifyReport", NotifyReportActivity, input)
}
