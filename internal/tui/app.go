package tui

import (
	"context"
	"fmt"

	graph "github.com/PizenLabs/lea/internal/graph/contracts"
	"github.com/PizenLabs/lea/internal/storage/contracts"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
)

type item struct {
	node *graph.Node
}

func (i item) Title() string { return i.node.Name }
func (i item) Description() string {
	return fmt.Sprintf("%s | %s:%d", i.node.Type, i.node.File, i.node.Line)
}
func (i item) FilterValue() string { return i.node.Name }

type model struct {
	store contracts.Store
	list  list.Model
}

func NewModel(store contracts.Store, nodes []*graph.Node) model {
	var items []list.Item
	for _, n := range nodes {
		items = append(items, item{node: n})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Symbols"

	return model{
		store: store,
		list:  l,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func Start(store contracts.Store) error {
	ctx := context.Background()
	nodes, err := store.ListNodes(ctx)
	if err != nil {
		return err
	}

	p := tea.NewProgram(NewModel(store, nodes), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
