package slack

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMessageWithClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      SendMessageInput
		handler    http.HandlerFunc
		wantSent   bool
		wantErr    bool
	}{
		{
			name: "successful send with text",
			input: SendMessageInput{
				Text: "hello world",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if len(body) == 0 {
					t.Error("expected non-empty body")
				}
				w.WriteHeader(http.StatusOK)
			},
			wantSent: true,
		},
		{
			name: "successful send with blocks",
			input: SendMessageInput{
				Text: "fallback",
				Blocks: []Block{
					{Type: BlockTypeHeader, Text: &Text{Type: TextTypePlainText, Text: "Title"}},
					{Type: BlockTypeDivider},
					{Type: BlockTypeSection, Text: &Text{Type: TextTypeMrkdwn, Text: "body"}},
				},
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantSent: true,
		},
		{
			name:     "empty text and blocks is noop",
			input:    SendMessageInput{},
			wantSent: false,
		},
		{
			name: "webhook error",
			input: SendMessageInput{
				Text: "hello",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("internal error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var webhookURL string
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				t.Cleanup(srv.Close)
				webhookURL = srv.URL
			}

			input := tt.input
			input.WebhookURL = webhookURL

			got, err := sendMessageWithClient(context.Background(), http.DefaultClient, input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("sendMessageWithClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.Sent != tt.wantSent {
				t.Errorf("Sent = %v, want %v", got.Sent, tt.wantSent)
			}
		})
	}
}
