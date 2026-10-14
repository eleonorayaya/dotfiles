package bar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const (
	utenaItem       = "utena"
	utenaRetryDelay = 5 * time.Second
)

type utenaSession struct {
	Name       string    `json:"name"`
	IsAttached bool      `json:"is_attached"`
	LastUsedAt time.Time `json:"last_used_at"`
}

type Utena struct {
	baseURL string
	client  *http.Client
}

func NewUtena(baseURL string) *Utena {
	return &Utena{baseURL: baseURL, client: &http.Client{Timeout: 60 * time.Second}}
}

func (u *Utena) Setup(ctx context.Context, b *Batch) error {
	b.Add("item", utenaItem, "right", P("icon", iconSession), "drawing=off")
	return nil
}

func (u *Utena) Stream(ctx context.Context, updates chan<- Update) {
	for ctx.Err() == nil {
		name, err := u.activeSession(ctx)
		if err != nil {
			slog.Debug("bar: utena unavailable", "error", err)
		}
		send(ctx, updates, func(b *Batch) {
			b.Set(utenaItem, onOff("drawing", name != ""), P("label", name))
		})
		if err != nil {
			sleep(ctx, utenaRetryDelay)
			continue
		}
		if err := u.waitForActivation(ctx); err != nil {
			slog.Debug("bar: utena activation poll failed", "error", err)
			sleep(ctx, utenaRetryDelay)
		}
	}
}

func (u *Utena) activeSession(ctx context.Context) (string, error) {
	var body struct {
		Sessions []utenaSession `json:"sessions"`
	}
	if err := u.get(ctx, "/sessions", &body); err != nil {
		return "", err
	}
	return pickActiveSession(body.Sessions), nil
}

func (u *Utena) waitForActivation(ctx context.Context) error {
	return u.get(ctx, "/tmux/activations/next", nil)
}

func (u *Utena) get(ctx context.Context, path string, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to build utena request: %w", err)
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("utena request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("utena %s: %s", path, resp.Status)
	}
	if into == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		return fmt.Errorf("failed to decode utena %s: %w", path, err)
	}
	return nil
}

func pickActiveSession(sessions []utenaSession) string {
	var best *utenaSession
	for i := range sessions {
		s := &sessions[i]
		if s.IsAttached && (best == nil || s.LastUsedAt.After(best.LastUsedAt)) {
			best = s
		}
	}
	if best == nil {
		return ""
	}
	return best.Name
}

func send(ctx context.Context, updates chan<- Update, u Update) {
	select {
	case updates <- u:
	case <-ctx.Done():
	}
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
