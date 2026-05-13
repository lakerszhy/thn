package ui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/dustin/go-humanize"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type itemsView struct {
	theme      Theme
	category   domain.Category
	pagination domain.Pagination
	client     *hn.Client

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

	return &itemsView{
		category:   category,
		client:     client,
		pagination: domain.NewPagination(),
		theme:      theme,
		model:      model,
	}
}

func (t *itemsView) Init() tea.Cmd {
	return func() tea.Msg {
		items, err := t.client.FetchItems(context.Background(), t.category, t.pagination)
		if err != nil {
			return err
		}
		return itemsMsg{
			category: t.category,
			items:    items,
		}
	}
}

func (t *itemsView) Update(msg tea.Msg) (*itemsView, tea.Cmd) {
	switch msg := msg.(type) {
	case itemsMsg:
		if msg.category == t.category {
			items := make([]list.Item, len(msg.items))
			for i, v := range msg.items {
				items[i] = listItem{item: v}
			}
			t.model.SetItems(items)
		}
		return t, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			index := t.model.Index()
			if index < 0 || index >= len(t.model.Items()) {
				return t, nil
			}

			if item, ok := t.model.Items()[index].(listItem); ok {
				return t, func() tea.Msg {
					return itemSelectedMsg(item.item)
				}
			}
		}
	}

	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t *itemsView) View() string {
	return t.model.View()
}

func (t *itemsView) setSize(width int, height int) {
	t.model.SetWidth(width)
	t.model.SetHeight(height)
}

type itemDeletage struct {
	theme Theme

	normalTitle   lipgloss.Style
	selectedTitle lipgloss.Style
	normalDesc    lipgloss.Style
	selectedDesc  lipgloss.Style

	ellipsis string
}

func newItemDeletage(t Theme) *itemDeletage {
	// 6: 1 for ">", 3 for index, 1 for ".", 1 for space
	desc := lipgloss.NewStyle().Padding(0, 0, 0, 6)
	return &itemDeletage{
		// 1 for ">"
		normalTitle:   lipgloss.NewStyle().Padding(0, 0, 0, 1).Foreground(t.itemTitleColor),
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

	if index == m.Index() {
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

func (d itemDeletage) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

type listItem struct {
	item domain.Item
}

func (listItem) FilterValue() string {
	return ""
}

func (l listItem) Description() string {
	v := fmt.Sprintf("%d points by %s %s",
		l.item.Score, l.item.By, humanize.Time(time.Unix(l.item.Time, 0)))

	if l.item.Descendants == 1 {
		v = fmt.Sprintf("%s | 1 comment", v)
	} else if l.item.Descendants > 1 {
		v = fmt.Sprintf("%s | %d comments", v, l.item.Descendants)
	}

	return v
}
