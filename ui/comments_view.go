package ui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type commentsView struct {
	item     domain.Item
	comments []domain.Item
	client   *hn.Client
}

func newCommentsView(item domain.Item, client *hn.Client) *commentsView {
	return &commentsView{
		item:   item,
		client: client,
	}
}

func (c *commentsView) Init() tea.Cmd {
	return func() tea.Msg {
		comments, err := c.client.FetchComments(context.Background(), c.item)
		if err != nil {
			return err
		}
		return commentsMsg{
			item:     c.item,
			comments: comments,
		}
	}
}

func (c *commentsView) Update(msg tea.Msg) (*commentsView, tea.Cmd) {
	switch msg := msg.(type) {
	case commentsMsg:
		if msg.item.ID == c.item.ID {
			c.comments = msg.comments
		}
		return c, nil
	}
	return c, nil
}

func (c *commentsView) View() string {
	var s strings.Builder

	fmt.Fprintf(&s, "%s\n\n", c.item.Title)

	for _, v := range c.comments {
		fmt.Fprintf(&s, "%s%s\n\n\n", v.Title, v.Text)
	}
	return s.String()
}
