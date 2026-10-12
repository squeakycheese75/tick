package cli

import (
	"context"
	"github.com/squeakycheese75/tick/internal/app"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCheckLoopbackAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		addr    string
		wantErr bool
	}{
		{"127.0.0.1:8080", false},
		{"localhost:8080", false},
		{"LocalHost:9000", false},
		{"[::1]:8080", false},
		{"127.0.0.2:8080", false},
		{"0.0.0.0:8080", true},
		{":8080", true},
		{"[::]:8080", true},
		{"192.168.1.20:8080", true},
		{"example.com:8080", true},
		{"8080", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			t.Parallel()

			err := checkLoopbackAddr(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkLoopbackAddr(%q) error = %v, wantErr %v", tt.addr, err, tt.wantErr)
			}
		})
	}
}

func TestCheckServeSettings_NeedsAStrongEnoughToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"none", "", true},
		{"only spaces", "                    ", true},
		{"too short", "abc123", true},
		{"one short of enough", strings.Repeat("a", minTokenLength-1), true},
		{"long enough", strings.Repeat("a", minTokenLength), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := checkServeSettings(defaultServeAddr, tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "TICK_API_TOKEN") {
				t.Fatalf("the message should say which setting to change: %v", err)
			}
		})
	}

	if err := checkServeSettings("0.0.0.0:8080", strings.Repeat("a", 32)); err == nil {
		t.Fatal("a good token must not make a non-loopback address acceptable")
	}
}

func TestServeCommand_RefusesToStartWithoutProtectionAndNeverBuildsTheRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TICK_API_TOKEN", "")

	built := false
	cmd := newServeCmd(func() (*app.Runtime, error) {
		built = true
		return nil, nil
	})
	cmd.SetArgs(nil)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := cmd.Execute()

	if err == nil || !strings.Contains(err.Error(), "TICK_API_TOKEN") {
		t.Fatalf("err = %v, want the token message", err)
	}
	if built {
		t.Fatal("the database must not be opened before the settings are checked")
	}
}

func TestServeCommand_RefusesANonLoopbackAddress(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TICK_API_TOKEN", strings.Repeat("t", 32))

	cmd := newServeCmd(func() (*app.Runtime, error) { t.Fatal("must not build the runtime"); return nil, nil })
	cmd.SetArgs([]string{"--addr", "0.0.0.0:8080"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

func TestServeUntilDone_ServesThenStopsCleanly(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "hello") })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() { done <- serveUntilDone(ctx, ln, handler) }()

	resp, err := http.Get("http://" + ln.Addr().String())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if string(body) != "hello" {
		t.Fatalf("body = %q", body)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("a requested stop is not an error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the server did not stop")
	}

	if _, err := http.Get("http://" + ln.Addr().String()); err == nil {
		t.Fatal("the port should be closed after shutdown")
	}
}

func TestServeUntilDone_ReportsAListenerThatFails(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	_ = ln.Close() // serving on a closed listener fails at once

	err = serveUntilDone(context.Background(), ln, http.NotFoundHandler())
	if err == nil {
		t.Fatal("expected the serve error to come back")
	}
}
