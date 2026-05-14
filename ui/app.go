package ui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type app struct {
	categories       []domain.Category
	current          domain.Category
	views            map[domain.Category]*itemsView
	commentsView     *commentsView
	focusOnItemsView bool

	client *hn.Client

	theme          Theme
	itemsViewStyle lipgloss.Style
	windowWidth    int
	windowHeight   int
}

func NewApp(client *hn.Client, theme Theme) *app {
	return &app{
		client:         client,
		theme:          theme,
		itemsViewStyle: lipgloss.NewStyle().Border(theme.border.style),
		categories: []domain.Category{
			domain.CategoryTop,
			domain.CategoryNew,
			domain.CategoryBest,
			domain.CategoryAsk,
			domain.CategoryShow,
			domain.CategoryJob,
		},
		focusOnItemsView: true,
		current:          domain.CategoryTop,
		views: map[domain.Category]*itemsView{
			domain.CategoryTop: newItemsView(domain.CategoryTop, client, theme),
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
		// TODO: how to handle when item has no comments
		a.focusOnItemsView = false
		a.commentsView = newCommentsView(domain.Item(msg), a.client, a.theme)
		a.updateSize()
		return a, a.commentsView.Init()
	case tea.WindowSizeMsg:
		a.windowWidth = msg.Width
		a.windowHeight = msg.Height
		a.itemsViewStyle = a.itemsViewStyle.Height(a.windowHeight).Width(a.windowWidth)
		a.updateSize()
		return a, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			a.focusOnItemsView = true
			a.commentsView = nil
			a.updateSize()
			return a, nil
		case "right", "l", "tab":
			index := slices.Index(a.categories, a.current)
			index = min(index+1, len(a.categories)-1)
			return a, a.updateCurrentCategory(index)
		case "left", "h", "shift+tab":
			index := slices.Index(a.categories, a.current)
			index = max(index-1, 0)
			return a, a.updateCurrentCategory(index)
		case "ctrl+c":
			return a, tea.Quit
		}

		if a.focusOnItemsView {
			for i := range a.views {
				if i == a.current {
					a.views[i], cmd = a.views[i].Update(msg)
					return a, cmd
				}
			}
			return a, nil
		} else if a.commentsView != nil {
			a.commentsView, cmd = a.commentsView.Update(msg)
			return a, cmd
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
	v.WindowTitle = "THN - Terminal for Hacker News"
	v.AltScreen = true

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		a.renderCategories(),
		a.views[a.current].View(),
	)

	style := a.itemsViewStyle.BorderForeground(a.theme.border.activeColor)
	if !a.focusOnItemsView {
		style = a.itemsViewStyle.BorderForeground(a.theme.border.color)
	}
	content = style.Render(content)

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

func (a *app) updateCurrentCategory(index int) tea.Cmd {
	a.current = a.categories[index]

	if _, ok := a.views[a.current]; !ok {
		a.views[a.current] = newItemsView(a.current, a.client, a.theme)
		a.updateSize()
		return a.views[a.current].Init()
	}

	return nil
}

func (a *app) updateSize() {
	// 2 for category bar
	contentHeight := a.windowHeight - a.itemsViewStyle.GetVerticalBorderSize() - 2
	contentWidth := a.windowWidth

	if a.commentsView != nil {
		contentWidth /= 2
		a.commentsView.setSize(a.windowWidth-contentWidth, a.windowHeight)
	}

	a.itemsViewStyle = a.itemsViewStyle.Width(contentWidth)

	for _, v := range a.views {
		v.setSize(contentWidth-a.itemsViewStyle.GetHorizontalFrameSize(), contentHeight)
	}
}

func (a app) renderCategories() string {
	catStyle := lipgloss.NewStyle().Padding(0, 1)

	categories := make([]string, len(a.categories))
	for i, c := range a.categories {
		if c == a.current {
			catStyle = catStyle.Foreground(a.theme.categoryActiveColor).Bold(true)
		} else {
			catStyle = catStyle.Foreground(a.theme.categoryColor)
		}
		categories[i] = catStyle.Render(string(c))
	}

	style := lipgloss.NewStyle().BorderForeground(a.theme.border.color).
		Border(a.theme.border.style, false, false, true, false).
		Width(a.itemsViewStyle.GetWidth() - a.itemsViewStyle.GetHorizontalFrameSize())
	if a.focusOnItemsView {
		style = style.BorderForeground(a.theme.border.activeColor)
	}

	return style.Render(strings.Join(categories, "|"))
}
