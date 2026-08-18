package mockapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type webhookSender struct {
	url      string
	client   *http.Client
	attempts int
	backoff  time.Duration
}

func newWebhookSender(url string, client *http.Client, attempts int, backoff time.Duration) webhookSender {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	if attempts < 1 {
		attempts = 3
	}
	if backoff <= 0 {
		backoff = 100 * time.Millisecond
	}

	return webhookSender{
		url:      url,
		client:   client,
		attempts: attempts,
		backoff:  backoff,
	}
}

func (s webhookSender) Deliver(ctx context.Context, event WebhookEvent) (int, error) {
	if s.url == "" {
		return 0, nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return 0, fmt.Errorf("marshal webhook: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= s.attempts; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
		if err != nil {
			return attempt - 1, fmt.Errorf("create webhook request: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")

		response, err := s.client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
			closeErr := response.Body.Close()
			if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
				if closeErr != nil {
					return attempt, fmt.Errorf("close webhook response: %w", closeErr)
				}
				return attempt, nil
			}
			lastErr = fmt.Errorf("webhook returned HTTP %d", response.StatusCode)
		} else {
			lastErr = err
		}

		if attempt == s.attempts {
			break
		}

		timer := time.NewTimer(s.backoff * time.Duration(attempt))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return attempt, errors.Join(lastErr, ctx.Err())
		case <-timer.C:
		}
	}

	return s.attempts, fmt.Errorf("webhook delivery failed after %d attempts: %w", s.attempts, lastErr)
}
