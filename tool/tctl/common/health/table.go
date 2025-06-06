// Teleport
// Copyright (C) 2025 Gravitational, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package health

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/trace"
)

type model struct {
	ctx            context.Context
	table          table.Model
	tableType      string
	clt            services.ProcessHealth
	selectedHostID string
}

func initHostsTable(ctx context.Context, clt services.ProcessHealth) (tea.Model, error) {
	rows, err := fetchHosts(ctx, clt)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	columns := []table.Column{
		{Title: "HostID", Width: 12},
		{Title: "Hostname", Width: 35},
		{Title: "Version", Width: 10},
		{Title: "Uptime", Width: 15},
		{Title: "Services (ok/total)", Width: 20},
		{Title: "_", Width: 2},
	}

	return model{
		ctx:       ctx,
		table:     tableWithColumnsRows(columns, rows),
		clt:       clt,
		tableType: "hosts",
	}, nil
}

func fetchHosts(ctx context.Context, clt services.ProcessHealth) ([]table.Row, error) {
	phs, _, err := clt.ListProcessHealths(ctx, 0, "")
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var hosts []table.Row
	for _, ph := range phs {
		uptime := ph.Status.SystemInfo.ProcessUptime
		uptimeSince := time.Now().Add(-time.Second * time.Duration(uptime))

		totalUnits := len(ph.Status.UnitsByName)
		okUnits := 0
		for _, unit := range ph.Status.UnitsByName {
			if unit.State == "ok" {
				okUnits++
			}
		}

		globalStatus := "✅"
		if totalUnits != okUnits {
			globalStatus = "⚠️"
		}

		hosts = append(hosts, table.Row{
			ph.Metadata.Name,
			ph.Status.SystemInfo.Hostname,
			ph.Status.SystemInfo.TeleportVersion,
			humanize.RelTime(uptimeSince, time.Now(), "ago", "from now"),
			fmt.Sprintf("%d/%d", okUnits, totalUnits),
			globalStatus,
		})
	}

	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i][1] < hosts[j][1] // Sort by hostname
	})

	return hosts, nil
}

func initUnitsTable(ctx context.Context, clt services.ProcessHealth, hostID string) (tea.Model, error) {
	rows, err := fetchUnits(ctx, clt, hostID)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	columns := []table.Column{
		{Title: "Unit", Width: 30},
		{Title: "State", Width: 5},
	}

	return model{
		ctx:            ctx,
		table:          tableWithColumnsRows(columns, rows),
		tableType:      "units",
		selectedHostID: hostID,
		clt:            clt,
	}, nil
}

func fetchUnits(ctx context.Context, clt services.ProcessHealth, hostID string) ([]table.Row, error) {
	var rows []table.Row

	phs, _, err := clt.ListProcessHealths(ctx, 0, "")
	if err != nil {
		return nil, trace.Wrap(err)
	}
	for _, ph := range phs {
		if ph.Metadata.Name != hostID {
			continue // Skip if the host ID does not match
		}

		for unitName, unit := range ph.Status.UnitsByName {
			rows = append(rows, table.Row{
				unitName,
				unit.State,
			})
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0] // Sort by unit name
	})

	return rows, nil
}

func tableWithColumnsRows(columns []table.Column, rows []table.Row) table.Model {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	t.SetStyles(s)

	return t
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			if m.tableType == "units" {
				newM, err := initUnitsTable(m.ctx, m.clt, m.selectedHostID)
				if err != nil {
					return m, tea.Quit
				}
				return newM, nil
			}

			newM, err := initHostsTable(m.ctx, m.clt)
			if err != nil {
				return m, tea.Quit
			}

			return newM, nil

		case "right":
			if m.tableType == "units" {
				return m, nil // No action if not in units view
			}

			newM, err := initUnitsTable(m.ctx, m.clt, m.table.SelectedRow()[0])
			if err != nil {
				return m, tea.Quit
			}

			return newM, nil
		case "left":
			if m.tableType == "hosts" {
				return m, nil // No action if not in units view
			}
			newM, err := initHostsTable(m.ctx, m.clt)
			if err != nil {
				return m, tea.Quit
			}

			return newM, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).Render(m.table.View()) + "\n"
}
