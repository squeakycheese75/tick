package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/delivery/api"
)

const (
	defaultServeAddr = "127.0.0.1:8080"
	minTokenLength   = 16
	shutdownGrace    = 10 * time.Second
)

func newServeCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the API over HTTP on this machine",
		Long: "Serve the same use cases the CLI runs over HTTP, for scripts and other programs on this machine.\n\n" +
			"It listens on the loopback interface only and requires a bearer token: set TICK_API_TOKEN\n" +
			"(in the environment or ~/.tick/config.env) to a random value of at least 16 characters, e.g.\n" +
			"  openssl rand -hex 32\n" +
			"See docs/openapi.yaml for the endpoints.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.LoadConfig()
			if err != nil {
				return err
			}

			if err := checkServeSettings(addr, cfg.APIToken); err != nil {
				return err
			}

			rt, err := runtimeBuilder()
			if err != nil {
				return err
			}

			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", addr, err)
			}

			handler := api.New(api.Deps{
				ListPortfolios:  rt.ListPortfolios,
				CreatePortfolio: rt.CreatePortfolio,
				ImportTrades:    rt.ImportTrades,
				GetSummary:      rt.GetPortfolioSummary,
				GetDaily:        rt.GetDailyReport,
			}, api.Options{
				Token:  cfg.APIToken,
				Logger: slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), nil)),
			})

			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Listening on http://%s (Ctrl-C to stop)\n", ln.Addr()); err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			return serveUntilDone(ctx, ln, handler)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", defaultServeAddr, "Loopback address to listen on")

	return cmd
}

// checkServeSettings stops the server starting somewhere it should not, or without protection.
func checkServeSettings(addr, token string) error {
	if err := checkLoopbackAddr(addr); err != nil {
		return err
	}

	if len(strings.TrimSpace(token)) < minTokenLength {
		return fmt.Errorf(
			"TICK_API_TOKEN must be set to a random value of at least %d characters (for example: openssl rand -hex 32)",
			minTokenLength,
		)
	}

	return nil
}

// checkLoopbackAddr refuses any address that is reachable from another machine. An empty host
// or 0.0.0.0 means every interface, which would expose the portfolio to the network.
func checkLoopbackAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid --addr %q: %w", addr, err)
	}

	if strings.EqualFold(host, "localhost") {
		return nil
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return nil
	}

	return fmt.Errorf("refusing to listen on %q: the API only serves this machine, so use 127.0.0.1, ::1 or localhost", addr)
}

// serveUntilDone serves until ctx is cancelled, then lets requests in flight finish.
func serveUntilDone(ctx context.Context, ln net.Listener, handler http.Handler) error {
	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	failed := make(chan error, 1)

	go func() { failed <- srv.Serve(ln) }()

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shut down: %w", err)
	}

	if err := <-failed; !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
