package ui

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
	"github.com/lakerszhy/thn/ui/comments"
	"github.com/lakerszhy/thn/ui/items"
)

type app struct {
	categories []domain.Category
	current    domain.Category

	titleBar     *titleBar
	itemsViews   map[domain.Category]*items.View
	commentsView *comments.View

	focusOnItemsView    bool
	commtentsFullscreen bool

	client *hn.Client

	theme  config.Theme
	hotkey config.Hotkey

	itemsViewStyle    lipgloss.Style
	commentsViewStyle lipgloss.Style

	windowWidth  int
	windowHeight int
}

func NewApp(client *hn.Client, theme config.Theme, hotkey config.Hotkey) tea.Model {
	categories := []domain.Category{
		domain.CategoryTop,
		domain.CategoryNew,
		domain.CategoryBest,
		domain.CategoryAsk,
		domain.CategoryShow,
		domain.CategoryJob,
	}

	itemsViewStyle := lipgloss.NewStyle().Border(theme.TitleBar.Border.Style).
		BorderForeground(theme.TitleBar.Border.Color)
	commentsViewStyle := lipgloss.NewStyle().Border(theme.Comment.Border.Style).
		BorderForeground(theme.Comment.Border.Color)

	return &app{
		client:            client,
		theme:             theme,
		hotkey:            hotkey,
		categories:        categories,
		current:           domain.CategoryTop,
		titleBar:          newTitleBar(categories, theme.TitleBar),
		focusOnItemsView:  true,
		itemsViewStyle:    itemsViewStyle,
		commentsViewStyle: commentsViewStyle,
		itemsViews: map[domain.Category]*items.View{
			domain.CategoryTop: items.NewView(domain.CategoryTop, client, theme, hotkey),
		},
	}
}

func (a *app) Init() tea.Cmd {
	a.titleBar.setCategory(a.current)

	return a.itemsViews[a.current].Init()
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case items.ItemSelectedMsg:
		a.focusOnItemsView = false
		a.commtentsFullscreen = msg.Fullscreen
		a.commentsView = comments.NewView(msg.Item.ID, a.client, a.theme, a.hotkey)
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

	var content string

	if !a.commtentsFullscreen {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			a.renderTitleBar(),
			a.itemsViews[a.current].View(),
		)

		style := a.itemsViewStyle
		if a.focusOnItemsView {
			style = style.BorderForeground(a.theme.TitleBar.Border.FocusColor)
		}
		content = style.Render(content)
	}

	if a.commentsView != nil {
		commentsStyle := a.commentsViewStyle
		if !a.focusOnItemsView {
			commentsStyle = commentsStyle.BorderForeground(a.theme.Comment.Border.FocusColor)
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
	a.titleBar.setCategory(a.current)

	if _, ok := a.itemsViews[a.current]; !ok {
		a.itemsViews[a.current] = items.NewView(a.current, a.client, a.theme, a.hotkey)
		a.updateSize()
		return a.itemsViews[a.current].Init()
	}

	return nil
}

func (a *app) updateSize() {
	itemsWidth := a.windowWidth

	if a.commtentsFullscreen {
		itemsWidth = 0
	}

	if a.commentsView != nil {
		itemsWidth /= 3
		commentsWidth := a.windowWidth - itemsWidth

		a.commentsViewStyle = a.commentsViewStyle.Width(commentsWidth)
		commentsInnerWidth := commentsWidth - a.commentsViewStyle.GetHorizontalFrameSize()
		commentsInnerHeight := a.windowHeight - a.commentsViewStyle.GetVerticalFrameSize()
		a.commentsView.SetSize(commentsInnerWidth, commentsInnerHeight)
	}

	a.itemsViewStyle = a.itemsViewStyle.Width(itemsWidth)

	availableWidth := itemsWidth - a.itemsViewStyle.GetHorizontalFrameSize()
	//nolint:mnd // 2 for title bar
	availableHeight := a.windowHeight - a.itemsViewStyle.GetVerticalBorderSize() - 2

	a.titleBar.setWidth(availableWidth)

	for _, v := range a.itemsViews {
		v.SetSize(availableWidth, availableHeight)
	}
}

func (a *app) renderTitleBar() string {
	style := lipgloss.NewStyle().Width(a.titleBar.width).
		Border(a.theme.TitleBar.Border.Style, false, false, true, false)

	if a.focusOnItemsView {
		style = style.BorderForeground(a.theme.TitleBar.Border.FocusColor)
	}

	return style.Render(a.titleBar.View())
}

func (a *app) onKeyPressMsg(msg tea.KeyPressMsg) (*app, tea.Cmd) {
	key := msg.String()

	if slices.Contains(a.hotkey.Quit, key) {
		return a, tea.Quit
	}

	if slices.Contains(a.hotkey.CloseComments, key) {
		a.focusOnItemsView = true
		a.commentsView = nil
		a.commtentsFullscreen = false
		a.updateSize()
		return a, nil
	}

	if !a.focusOnItemsView && slices.Contains(a.hotkey.ToggleFullscreen, key) {
		a.commtentsFullscreen = !a.commtentsFullscreen
		a.updateSize()
		return a, nil
	}

	var cmd tea.Cmd

	if a.commentsView != nil && !a.focusOnItemsView {
		a.commentsView, cmd = a.commentsView.Update(msg)
		return a, cmd
	}

	if slices.Contains(a.hotkey.NextCategory, key) {
		index := slices.Index(a.categories, a.current)
		index = min(index+1, len(a.categories)-1)
		return a, a.updateCurrentCategory(index)
	}

	if slices.Contains(a.hotkey.PrevCategory, key) {
		index := slices.Index(a.categories, a.current)
		index = max(index-1, 0)
		return a, a.updateCurrentCategory(index)
	}

	a.itemsViews[a.current], cmd = a.itemsViews[a.current].Update(msg)
	return a, cmd
}
