package ui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type itemsView struct {
	category   domain.Category
	pagination domain.Pagination
	client     *hn.Client
	items      []domain.Item
}

func newItemsView(category domain.Category, client *hn.Client) *itemsView {
	return &itemsView{
		category:   category,
		client:     client,
		pagination: domain.NewPagination(),
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

	return t, nil
}

func (t *itemsView) View() string {
	var s strings.Builder

	for i, v := range t.items {
		fmt.Fprintf(&s, "%d. %s\n", i+1, v.Title)
	}

	return s.String()
}
