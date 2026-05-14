package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type itemsView struct {
	theme      Theme
	category   domain.Category
	pagination domain.Pagination
	client     *hn.Client
	msg        itemsMsg
	spinner    spinner.Model

	model  list.Model
	width  int
	height int
}

func newItemsView(category domain.Category, client *hn.Client, theme Theme) *itemsView {
	model := list.New(nil, newItemDeletage(theme), 0, 0)
	model.SetShowTitle(false)
	model.SetFilteringEnabled(false)
	model.SetShowStatusBar(false)
	model.SetShowPagination(false)
	model.SetShowHelp(false)
	model.DisableQuitKeybindings()

	s := spinner.New()
	s.Spinner = spinner.Dot

	return &itemsView{
		category:   category,
		client:     client,
		pagination: domain.NewPagination(),
		theme:      theme,
		model:      model,
		spinner:    s,
	}
}

func (t *itemsView) Init() tea.Cmd {
	return tea.Batch(
		t.spinner.Tick,
		t.fetch(),
	)
}

func (t *itemsView) Update(msg tea.Msg) (*itemsView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if t.msg.isLoading() {
			t.spinner, cmd = t.spinner.Update(msg)
		}
		return t, cmd
	case itemsMsg:
		if msg.category != t.category {
			return t, nil
		}

		t.msg = msg

		switch msg.state {
		case stateLoading, stateLoadFailed:
			t.model.SetItems(nil)
			return t, nil
		case stateLoadSuccess:
			items := make([]list.Item, len(msg.items)+1)
			for i, v := range msg.items {
				items[i] = listItem{item: v}
			}
			items[len(items)-1] = loadMoreItem{}
			t.model.SetItems(items)
		case stateLoadMoreSuccess:
			t.pagination = t.pagination.Next()

			items := make([]list.Item, 0, len(msg.items)+len(t.msg.items))
			items = append(items, t.model.Items()[0:len(t.model.Items())-1]...)
			for _, v := range msg.items {
				items = append(items, listItem{item: v})
			}
			items = append(items, loadMoreItem{})
			t.model.SetItems(items)
		}
		// handle by delegate, so not return
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			index := t.model.Index()
			if index < 0 || index >= len(t.model.Items()) {
				return t, nil
			}

			switch i := t.model.Items()[index].(type) {
			case listItem:
				return t, func() tea.Msg {
					return itemSelectedMsg(i.item)
				}
			case loadMoreItem:
				return t, t.fetchMore()
			}
		}
	}

	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t *itemsView) View() string {
	switch t.msg.state {
	case stateLoading:
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(t.model.Width()).
			Render(fmt.Sprintf("%s Loading...", t.spinner.View()))
	case stateLoadFailed:
		return fmt.Sprintf("Load Failed: %s", t.msg.err.Error())
	case stateLoadSuccess, stateLoadMoreSuccess, stateLoadMoreFailed:
		return t.model.View()
	}
	return t.model.View()
}

func (t *itemsView) setSize(width int, height int) {
	t.model.SetWidth(width)
	t.model.SetHeight(height)
}

func (t itemsView) fetch() tea.Cmd {
	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newItemsLoadingMsg(t.category)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := t.client.FetchItems(context.Background(), t.category, t.pagination)
		if err != nil {
			return newItemsLoadFailedMsg(t.category, err)
		}
		return newItemsLoadSuccessMsg(t.category, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (t itemsView) fetchMore() tea.Cmd {
	if t.msg.isLoading() || t.msg.isLoadingMore() {
		return nil
	}

	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newItemsLoadingMoreMsg(t.category)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := t.client.FetchItems(context.Background(), t.category, t.pagination.Next())
		if err != nil {
			return newItemsLoadMoreFailedMsg(t.category, err)
		}
		return newItemsLoadMoreSuccessMsg(t.category, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

type itemDeletage struct {
	theme Theme

	normalTitle   lipgloss.Style
	selectedTitle lipgloss.Style
	normalDesc    lipgloss.Style
	selectedDesc  lipgloss.Style

	ellipsis string

	msg itemsMsg
}

func newItemDeletage(t Theme) *itemDeletage {
	// 6: 1 for ">", 3 for index, 1 for ".", 1 for space
	desc := lipgloss.NewStyle().PaddingLeft(6)
	return &itemDeletage{
		// 1 for ">"
		normalTitle:   lipgloss.NewStyle().PaddingLeft(1).Foreground(t.itemTitleColor),
		normalDesc:    desc.Foreground(t.itemDescColor).Faint(true),
		selectedTitle: lipgloss.NewStyle().Foreground(t.itemTitleSelectedColor),
		selectedDesc:  desc.Foreground(t.itemDescSelectedColor).Faint(true),
		ellipsis:      "...",
	}
}

func (d itemDeletage) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if m.Width() <= 0 {
		return
	}

	var selected = index == m.Index()

	if _, ok := item.(loadMoreItem); ok {
		fmt.Fprintf(w, "%s", d.renderLoadMore(selected))
		return
	}

	var title, desc string

	if i, ok := item.(listItem); ok {
		title = fmt.Sprintf("%3d. %s", index+1, i.item.Title)
		desc = i.Description()
	} else {
		return
	}

	textwidth := m.Width() - d.normalTitle.GetHorizontalPadding()
	title = ansi.Truncate(title, textwidth, d.ellipsis)

	var lines []string
	for i, line := range strings.Split(desc, "\n") {
		if i >= d.Height()-1 {
			break
		}
		lines = append(lines, ansi.Truncate(line, textwidth, d.ellipsis))
	}
	desc = strings.Join(lines, "\n")

	if selected {
		title = d.selectedTitle.Render(">" + title)
		desc = d.selectedDesc.Render(desc)
	} else {
		title = d.normalTitle.Render(title)
		desc = d.normalDesc.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func (d itemDeletage) Height() int {
	return 2
}

func (d itemDeletage) Spacing() int {
	return 0
}

func (d *itemDeletage) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	switch msg := msg.(type) {
	case itemsMsg:
		d.msg = msg
	}
	return nil
}

func (d itemDeletage) renderLoadMore(selected bool) string {
	content := "More"
	switch d.msg.state {
	case stateLoadingMore:
		content = "Loading..."
	case stateLoadMoreFailed:
		content = fmt.Sprintf("Load more failed: %s", d.msg.err.Error())
	}

	if selected {
		content = lipgloss.NewStyle().PaddingLeft(1).Render(content)
		content = d.selectedTitle.Render(">" + content)
	} else {
		content = lipgloss.NewStyle().PaddingLeft(1).Render(content)
		content = d.normalTitle.Render(content)
	}

	return fmt.Sprintf("%s", content)
}

type listItem struct {
	item domain.Item
}

func (listItem) FilterValue() string {
	return ""
}

func (l listItem) Description() string {
	v := fmt.Sprintf("%d points by %s %s", l.item.Score, l.item.By, l.item.TimeAgo())

	if l.item.Descendants == 1 {
		v = fmt.Sprintf("%s | 1 comment", v)
	} else if l.item.Descendants > 1 {
		v = fmt.Sprintf("%s | %d comments", v, l.item.Descendants)
	}

	return v
}

type loadMoreItem struct {
}

func (l loadMoreItem) FilterValue() string {
	return ""
}
