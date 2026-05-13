package ui

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type app struct {
	categories   []domain.Category
	current      domain.Category
	views        map[domain.Category]*itemsView
	commentsView *commentsView
	client       *hn.Client
}

func NewApp(client *hn.Client) *app {
	return &app{
		client: client,
		categories: []domain.Category{
			domain.CategoryTop,
			domain.CategoryNew,
			domain.CategoryBest,
			domain.CategoryAsk,
			domain.CategoryShow,
			domain.CategoryJob,
		},
		current: domain.CategoryTop,
		views: map[domain.Category]*itemsView{
			domain.CategoryTop: newItemsView(domain.CategoryTop, client),
		},
	}
}

func (a *app) Init() tea.Cmd {
	return a.views[a.current].Init()
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case itemSelectedMsg:
		a.commentsView = newCommentsView(domain.Item(msg), a.client)
		return a, a.commentsView.Init()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "right", "l", "tab":
			index := slices.Index(a.categories, a.current)
			index = min(index+1, len(a.categories)-1)
			a.current = a.categories[index]
			if _, ok := a.views[a.current]; !ok {
				a.views[a.current] = newItemsView(a.current, a.client)
				cmd = a.views[a.current].Init()
			}
			return a, cmd
		case "left", "h", "shift+tab":
			index := slices.Index(a.categories, a.current)
			index = max(index-1, 0)
			a.current = a.categories[index]
			if _, ok := a.views[a.current]; !ok {
				a.views[a.current] = newItemsView(a.current, a.client)
				cmd = a.views[a.current].Init()
			}
			return a, cmd
		case "ctrl+c":
			return a, tea.Quit
		}

		for i := range a.views {
			if i == a.current {
				a.views[i], cmd = a.views[i].Update(msg)
				return a, cmd
			}
		}

		return a, nil
	}

	var cmds []tea.Cmd

	for i := range a.views {
		a.views[i], cmd = a.views[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	if a.commentsView != nil {
		a.commentsView, cmd = a.commentsView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a *app) View() tea.View {
	var v tea.View
	v.WindowTitle = "Hacker News"
	v.AltScreen = true

	style := lipgloss.NewStyle().Padding(0, 1)

	categories := make([]string, len(a.categories))
	for _, c := range a.categories {
		if c == a.current {
			style = style.Background(lipgloss.Red)
		} else {
			style = style.UnsetBackground()
		}
		categories = append(categories, style.Render(string(c)))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top, categories...,
		),
		a.views[a.current].View(),
	)

	if a.commentsView != nil {
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			content,
			a.commentsView.View(),
		)
	}

	v.Content = content
	return v
}
