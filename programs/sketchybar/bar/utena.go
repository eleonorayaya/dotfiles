package bar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	utenaItem       = "utena"
	utenaURL        = "http://localhost:3333"
	utenaRetryDelay = 5 * time.Second
	utenaPollWindow = 10 * time.Second
	utenaSettle     = time.Second
)

type utenaSession struct {
	ID             uint          `json:"ID"`
	Status         string        `json:"status"`
	ClaudeSessions []claudeAgent `json:"claude_sessions"`
	Name           string        `json:"name"`
	IsAttached     bool          `json:"is_attached"`
	LastUsedAt     time.Time     `json:"last_used_at"`
}

type Utena struct {
	utenaClient
	colors     Colors
	pollWindow time.Duration
	settle     time.Duration
	last       lastValue
}

func NewUtena(baseURL string, colors Colors) *Utena {
	return &Utena{
		utenaClient: newUtenaClient(baseURL),
		colors:      colors,
		pollWindow:  utenaPollWindow,
		settle:      utenaSettle,
	}
}

func (u *Utena) Setup(ctx context.Context, b *Batch) error {
	b.Add(utenaItem, "left", append(island(u.colors), "icon.drawing=off", P("label.padding_left", islandPadding), "drawing=off")...)
	return nil
}

func (u *Utena) Stream(ctx context.Context, updates chan<- Update) {
	for ctx.Err() == nil {
		name, err := u.activeSession(ctx)
		if err != nil {
			slog.Debug("bar: utena unavailable", "error", err)
		}
		if u.last.changed(name) {
			send(ctx, updates, func(b *Batch) {
				b.Set(utenaItem, onOff("drawing", name != ""), P("label", sparkleFrame(name)))
			})
		}
		if err != nil {
			sleep(ctx, utenaRetryDelay)
			continue
		}
		if err := u.waitForActivation(ctx); err != nil {
			slog.Debug("bar: utena activation poll failed", "error", err)
			sleep(ctx, utenaRetryDelay)
			continue
		}
		sleep(ctx, u.settle)
	}
}

func (u *Utena) activeSession(ctx context.Context) (string, error) {
	sessions, err := u.sessions(ctx)
	if err != nil {
		return "", err
	}
	return pickActiveSession(sessions), nil
}

func (u *Utena) waitForActivation(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, u.pollWindow)
	defer cancel()
	err := u.do(ctx, http.MethodGet, "/tmux/activations/next", nil)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil
	}
	return err
}

type utenaClient struct {
	baseURL string
	client  *http.Client
}

func newUtenaClient(baseURL string) utenaClient {
	return utenaClient{baseURL: baseURL, client: &http.Client{Timeout: 60 * time.Second}}
}

func (u utenaClient) do(ctx context.Context, method, path string, into any) error {
	req, err := http.NewRequestWithContext(ctx, method, u.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to build utena request: %w", err)
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("utena request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("utena %s %s: %s", method, path, resp.Status)
	}
	if into == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		return fmt.Errorf("failed to decode utena %s: %w", path, err)
	}
	return nil
}

func (u utenaClient) sessions(ctx context.Context) ([]utenaSession, error) {
	var body struct {
		Sessions []utenaSession `json:"sessions"`
	}
	if err := u.do(ctx, http.MethodGet, "/sessions?status="+strings.Join(liveSessionStatuses, ","), &body); err != nil {
		return nil, err
	}
	return body.Sessions, nil
}

func (u *Utena) resetCache() {
	u.last.reset()
}

func sparkleFrame(s string) string {
	return "˚ ₊ " + s + " ₊ ˚"
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
