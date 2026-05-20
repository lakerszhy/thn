package ui

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
	"github.com/lakerszhy/thn/ui/comments"
	"github.com/lakerszhy/thn/ui/items"
)

type app struct {
	categories       []domain.Category
	current          domain.Category
	itemsViews       map[domain.Category]*items.View
	commentsView     *comments.View
	focusOnItemsView bool

	client *hn.Client

	theme  config.Theme
	hotkey config.Hotkey

	itemsViewStyle    lipgloss.Style
	commentsViewStyle lipgloss.Style

	windowWidth  int
	windowHeight int
}

func NewApp(client *hn.Client, theme config.Theme, hotkey config.Hotkey) tea.Model {
	return &app{
		client:            client,
		theme:             theme,
		hotkey:            hotkey,
		itemsViewStyle:    lipgloss.NewStyle().Border(theme.TitleBar.Border.Style),
		commentsViewStyle: lipgloss.NewStyle().Border(theme.Comment.Border.Style),
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
		itemsViews: map[domain.Category]*items.View{
			domain.CategoryTop: items.NewView(domain.CategoryTop, client, theme, hotkey),
		},
	}
}

func (a *app) Init() tea.Cmd {
	return a.itemsViews[a.current].Init()
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case items.ItemSelectedMsg:
		// TODO: how to handle when item has no comments
		a.focusOnItemsView = false
		a.commentsView = comments.NewView(domain.Item(msg).ID, a.client, a.theme, a.hotkey)
		a.updateSize()
		return a, a.commentsView.Init()
	case tea.WindowSizeMsg:
		a.windowWidth = msg.Width
		a.windowHeight = msg.Height
		a.itemsViewStyle = a.itemsViewStyle.Height(a.windowHeight).Width(a.windowWidth)
		a.commentsViewStyle = a.commentsViewStyle.Height(a.windowHeight)
		a.updateSize()
		return a, nil
	case tea.KeyPressMsg:
		return a.onKeyPressMsg(msg)
	}

	var cmds []tea.Cmd

	for i := range a.itemsViews {
		a.itemsViews[i], cmd = a.itemsViews[i].Update(msg)
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
		a.itemsViews[a.current].View(),
	)

	style := a.itemsViewStyle.BorderForeground(a.theme.TitleBar.Border.FocusColor)
	if !a.focusOnItemsView {
		style = a.itemsViewStyle.BorderForeground(a.theme.TitleBar.Border.Color)
	}
	content = style.Render(content)

	if a.commentsView != nil {
		commentsStyle := a.commentsViewStyle.BorderForeground(a.theme.Comment.Border.Color)
		if !a.focusOnItemsView {
			commentsStyle = a.commentsViewStyle.BorderForeground(a.theme.Comment.Border.FocusColor)
		}

		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			content,
			commentsStyle.Render(a.commentsView.View()),
		)
	}

	v.Content = content
	return v
}

func (a *app) updateCurrentCategory(index int) tea.Cmd {
	a.current = a.categories[index]

	if _, ok := a.itemsViews[a.current]; !ok {
		a.itemsViews[a.current] = items.NewView(a.current, a.client, a.theme, a.hotkey)
		a.updateSize()
		return a.itemsViews[a.current].Init()
	}

	return nil
}

func (a *app) updateSize() {
	itemsWidth := a.windowWidth

	if a.commentsView != nil {
		itemsWidth /= 3
		commentsWidth := a.windowWidth - itemsWidth

		a.commentsViewStyle = a.commentsViewStyle.Width(commentsWidth)
		commentsInnerWidth := commentsWidth - a.commentsViewStyle.GetHorizontalFrameSize()
		commentsInnerHeight := a.windowHeight - a.commentsViewStyle.GetVerticalFrameSize()
		a.commentsView.SetSize(commentsInnerWidth, commentsInnerHeight)
	}

	a.itemsViewStyle = a.itemsViewStyle.Width(itemsWidth)

	//nolint:mnd // 2 for category bar
	itemsHeight := a.windowHeight - a.itemsViewStyle.GetVerticalBorderSize() - 2
	itemsWidth -= a.itemsViewStyle.GetHorizontalFrameSize()

	for _, v := range a.itemsViews {
		v.SetSize(itemsWidth, itemsHeight)
	}
}

func (a *app) renderCategories() string {
	catStyle := lipgloss.NewStyle().Padding(0, 1)

	categories := make([]string, len(a.categories))
	for i, c := range a.categories {
		if c == a.current {
			catStyle = catStyle.Foreground(a.theme.TitleBar.CategorySelectedColor).Bold(true)
		} else {
			catStyle = catStyle.Foreground(a.theme.TitleBar.CategoryColor)
		}
		categories[i] = catStyle.Render(string(c))
	}

	style := lipgloss.NewStyle().BorderForeground(a.theme.TitleBar.Border.Color).
		Border(a.theme.TitleBar.Border.Style, false, false, true, false).
		Width(a.itemsViewStyle.GetWidth() - a.itemsViewStyle.GetHorizontalFrameSize())
	if a.focusOnItemsView {
		style = style.BorderForeground(a.theme.TitleBar.Border.FocusColor)
	}

	divider := lipgloss.NewStyle().Foreground(a.theme.TitleBar.DivideColor).Render("|")

	return style.Render(strings.Join(categories, divider))
}

func (a *app) onKeyPressMsg(msg tea.KeyPressMsg) (*app, tea.Cmd) {
	if key.Matches(msg, a.hotkey.Quit) {
		return a, tea.Quit
	}

	if key.Matches(msg, a.hotkey.CloseComments) {
		a.focusOnItemsView = true
		a.commentsView = nil
		a.updateSize()
		return a, nil
	}

	var cmd tea.Cmd

	if a.focusOnItemsView {
		if key.Matches(msg, a.hotkey.NextCategory) {
			index := slices.Index(a.categories, a.current)
			index = min(index+1, len(a.categories)-1)
			return a, a.updateCurrentCategory(index)
		}

		if key.Matches(msg, a.hotkey.PrevCategory) {
			index := slices.Index(a.categories, a.current)
			index = max(index-1, 0)
			return a, a.updateCurrentCategory(index)
		}
	} else {
		a.commentsView, cmd = a.commentsView.Update(msg)
		return a, cmd
	}

	a.itemsViews[a.current], cmd = a.itemsViews[a.current].Update(msg)
	return a, cmd
}
