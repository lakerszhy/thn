package ui

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type itemsView struct {
	theme      config.Theme
	category   domain.Category
	pagination domain.Pagination
	client     *hn.Client
	msg        itemsMsg
	spinner    spinner.Model

	model  list.Model
	width  int
	height int
}

func newItemsView(category domain.Category, client *hn.Client, theme config.Theme) *itemsView {
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
				items[i] = listItem{Item: v}
			}
			items[len(items)-1] = loadMoreItem{}
			t.model.SetItems(items)
		case stateLoadMoreSuccess:
			t.pagination = t.pagination.Next()

			items := make([]list.Item, 0, len(msg.items)+len(t.msg.items))
			items = append(items, t.model.Items()[0:len(t.model.Items())-1]...)
			for _, v := range msg.items {
				items = append(items, listItem{Item: v})
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
					return itemSelectedMsg(i.Item)
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
	theme config.Theme

	normalTitle    lipgloss.Style
	selectedTitle  lipgloss.Style
	normalDesc     lipgloss.Style
	selectedDesc   lipgloss.Style
	normalDomain   lipgloss.Style
	selectedDomain lipgloss.Style

	ellipsis string

	msg itemsMsg
}

func newItemDeletage(t config.Theme) *itemDeletage {
	// 6: 1 for ">", 3 for index, 1 for ".", 1 for space
	desc := lipgloss.NewStyle().PaddingLeft(6)
	return &itemDeletage{
		// 1 for ">"
		normalTitle:    lipgloss.NewStyle().PaddingLeft(1).Foreground(t.Item.TitleColor),
		normalDesc:     desc.Foreground(t.Item.DescColor),
		selectedTitle:  lipgloss.NewStyle().Foreground(t.Item.TitleSelectedColor),
		selectedDesc:   desc.Foreground(t.Item.DescSelectedColor),
		normalDomain:   desc.Foreground(t.Item.DomainColor),
		selectedDomain: desc.Foreground(t.Item.DomainSelectedColor),
		ellipsis:       "...",
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

	var title, desc, domain string

	if i, ok := item.(listItem); ok {
		title = fmt.Sprintf("%3d. %s", index+1, i.Title)
		domain = i.Domain()
		desc = i.Description()
	} else {
		return
	}

	if domain != "" {
		domain = fmt.Sprintf(" (%s)", domain)
	}

	if selected {
		title = d.selectedTitle.Render(">" + title)
		domain = d.selectedDomain.UnsetPadding().Render(domain)
		desc = d.selectedDesc.Render(desc)
	} else {
		title = d.normalTitle.Render(title)
		domain = d.normalDomain.UnsetPadding().Render(domain)
		desc = d.normalDesc.Render(desc)
	}

	textwidth := m.Width() - d.normalTitle.GetHorizontalPadding()

	title = fmt.Sprintf("%s%s", title, domain)
	title = ansi.Truncate(title, textwidth, d.ellipsis)

	desc = ansi.Truncate(desc, textwidth, d.ellipsis)

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
	domain.Item
}

func (listItem) FilterValue() string {
	return ""
}

func (l listItem) Description() string {
	v := fmt.Sprintf("%d points by %s %s", l.Score, l.By, l.TimeAgo())

	if l.Descendants == 1 {
		v = fmt.Sprintf("%s | 1 comment", v)
	} else if l.Descendants > 1 {
		v = fmt.Sprintf("%s | %d comments", v, l.Descendants)
	}

	return v
}

func (l listItem) Domain() string {
	u, err := url.Parse(l.URL)
	if err != nil {
		return ""
	}

	host := strings.TrimPrefix(u.Hostname(), "www.")

	if host == "github.com" || host == "twitter.com" || host == "x.com" {
		paths := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(paths) > 1 {
			r, _ := url.JoinPath(host, paths[0])
			return r
		}
	}

	return host
}

type loadMoreItem struct {
}

func (l loadMoreItem) FilterValue() string {
	return ""
}
