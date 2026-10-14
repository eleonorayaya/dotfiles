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
	"strconv"
	"strings"
	"syscall"
	"time"
)

const takeoverTimeout = 2 * time.Second

var forwardedEnv = []string{"BUTTON", "MODIFIER", "FOCUSED_WORKSPACE", "PREV_WORKSPACE"}

type Event struct {
	Name   string            `json:"name"`
	Sender string            `json:"sender"`
	Info   string            `json:"info,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
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

type Server struct {
	net.Listener
	pidPath string
}

func Listen(socketPath string) (*Server, error) {
	pidPath := socketPath + ".pid"
	stopPrevious(socketPath, pidPath)

	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to remove stale socket: %w", err)
	}
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", socketPath, err)
	}
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		l.Close()
		return nil, fmt.Errorf("failed to write pidfile: %w", err)
	}
	return &Server{Listener: l, pidPath: pidPath}, nil
}

func (s *Server) Close() error {
	if pid, err := readPid(s.pidPath); err == nil && pid == os.Getpid() {
		os.Remove(s.pidPath)
	}
	return s.Listener.Close()
}

func stopPrevious(socketPath, pidPath string) {
	conn, err := net.DialTimeout("unix", socketPath, time.Second)
	if err != nil {
		return
	}
	conn.Close()

	pid, err := readPid(pidPath)
	if err != nil || pid == os.Getpid() {
		return
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return
	}
	deadline := time.Now().Add(takeoverTimeout)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	slog.Warn("bar: previous daemon ignored SIGTERM, killing", "pid", pid)
	syscall.Kill(pid, syscall.SIGKILL)
}

func readPid(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
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
