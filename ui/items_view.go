package ui

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/dustin/go-humanize"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type itemsView struct {
	theme      Theme
	category   domain.Category
	pagination domain.Pagination
	client     *hn.Client
	items      []domain.Item

	model  list.Model
	width  int
	height int
}

func newItemsView(category domain.Category, client *hn.Client, theme Theme) *itemsView {
	delegate := list.NewDefaultDelegate()
	model := list.New(nil, delegate, 0, 0)
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
			t.items = msg.items

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
			if len(t.items) > 0 {
				return t, func() tea.Msg {
					return itemSelectedMsg(t.items[0])
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

type listItem struct {
	item domain.Item
}

func (l listItem) FilterValue() string {
	return l.item.Title
}

func (l listItem) Title() string {
	return l.item.Title
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
