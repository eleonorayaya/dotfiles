package bar

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"time"
)

const quitSender = "shizuku.quit"

var forwardedEnv = []string{"NAME", "SENDER", "INFO", "BUTTON", "MODIFIER", "FOCUSED_WORKSPACE", "PREV_WORKSPACE"}

type Event struct {
	Name   string            `json:"name"`
	Sender string            `json:"sender"`
	Info   string            `json:"info,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
}

func DefaultSocketPath() string {
	return filepath.Join(os.TempDir(), "shizuku-bar.sock")
}

func EventFromEnv() Event {
	ev := Event{
		Name:   os.Getenv("NAME"),
		Sender: os.Getenv("SENDER"),
		Info:   os.Getenv("INFO"),
		Env:    map[string]string{},
	}
	for _, key := range forwardedEnv {
		if v, ok := os.LookupEnv(key); ok {
			ev.Env[key] = v
		}
	}
	return ev
}

func Emit(socketPath string, ev Event) error {
	conn, err := net.DialTimeout("unix", socketPath, time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to bar daemon: %w", err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(ev); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}
	return nil
}

func Listen(socketPath string) (net.Listener, error) {
	if _, err := os.Stat(socketPath); err == nil {
		if err := Emit(socketPath, Event{Sender: quitSender}); err == nil {
			waitForRemoval(socketPath, 2*time.Second)
		}
		if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to remove stale socket: %w", err)
		}
	}

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", socketPath, err)
	}
	return l, nil
}

func waitForRemoval(path string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func Serve(ctx context.Context, l net.Listener, events chan<- Event) {
	go func() {
		<-ctx.Done()
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() == nil && !errors.Is(err, net.ErrClosed) {
				slog.Warn("bar: accept failed", "error", err)
			}
			return
		}
		go func() {
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				var ev Event
				if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
					slog.Warn("bar: bad event", "error", err)
					continue
				}
				select {
				case events <- ev:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}
