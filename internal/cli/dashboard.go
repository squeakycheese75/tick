package cli

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/squeakycheese75/tick/internal/app"
	"github.com/squeakycheese75/tick/internal/domain"
)

func newDashboardCmd(runtimeBuilder RuntimeBuilder) *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard",
		Short: "Open interactive dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := runtimeBuilder()
			if err != nil {
				return err
			}

			return runDashboard(cmd.Context(), app)
		},
	}
}

func runDashboard(ctx context.Context, rt *app.Runtime) error {
	m := dashboardModel{
		ctx:     ctx,
		runtime: rt,
		loading: true,
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

type dashboardModel struct {
	ctx     context.Context
	runtime *app.Runtime

	report  domain.GetDailyReportOutput
	err     error
	loading bool
}

type dashboardLoadedMsg struct {
	report domain.GetDailyReportOutput
	err    error
}

func loadDashboardCmd(ctx context.Context, rt *app.Runtime) tea.Cmd {
	return func() tea.Msg {
		out, err := rt.GetDailyReport.Execute(ctx, domain.GetDailyReportInput{
			PortfolioName: "main",
			NewsLimit:     2,
		})
		return dashboardLoadedMsg{
			report: out,
			err:    err,
		}
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return loadDashboardCmd(m.ctx, m.runtime)
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case dashboardLoadedMsg:
		m.loading = false
		m.report = msg.report
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, loadDashboardCmd(m.ctx, m.runtime)
		}
	}

	return m, nil
}

func (m dashboardModel) View() string {
	if m.loading {
		return "Loading...\n\n[q] quit"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\n[r] retry  [q] quit", m.err)
	}

	r := m.report.DailyReport

	out := ""
	out += fmt.Sprintf("Portfolio: %s\n", r.PortfolioName)
	out += fmt.Sprintf("Total value: %.2f %s\n", r.TotalValue, r.BaseCurrency)

	if r.ChangeSinceLastSnapshot != nil {
		out += fmt.Sprintf(
			"Since last snapshot: %+.2f %s (%+.2f%%)\n",
			r.ChangeSinceLastSnapshot.Absolute,
			r.BaseCurrency,
			r.ChangeSinceLastSnapshot.Percent*100,
		)
	}

	out += "\nTop holdings\n"
	for _, h := range r.TopHoldings {
		out += fmt.Sprintf("- %s %.2f%%\n", h.Symbol, h.Weight*100)
	}

	out += "\n[r] refresh  [q] quit\n"

	return out
}
