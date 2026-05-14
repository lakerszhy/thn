package ui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type commentsView struct {
	item      domain.Item
	client    *hn.Client
	theme     Theme
	model     viewport.Model
	converter *converter.Converter
}

func newCommentsView(item domain.Item, client *hn.Client, theme Theme) *commentsView {
	converter := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
		),
	)
	vp := viewport.New()
	vp.SoftWrap = true
	return &commentsView{
		item:      item,
		client:    client,
		theme:     theme,
		model:     vp,
		converter: converter,
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
			comments := make([]domain.Item, 0, len(msg.comments))
			for i := range msg.comments {
				msg.comments[i].Text = c.renderComment(msg.comments[i])
				comments = append(comments, msg.comments[i])
			}

			var s strings.Builder
			for _, v := range comments {
				fmt.Fprintf(&s, "%s\n", v.Text)
			}

			c.model.SetContent(s.String())
			return c, nil
		}
		return c, nil
	}

	var cmd tea.Cmd
	c.model, cmd = c.model.Update(msg)
	return c, cmd
}

func (c *commentsView) View() string {
	return c.model.View()
}

func (c *commentsView) setSize(width, height int) {
	c.model.SetHeight(height)
	c.model.SetWidth(width)
}

func (c *commentsView) renderComment(comment domain.Item) string {
	desc := fmt.Sprintf("%s %s", comment.By, comment.TimeAgo())
	desc = lipgloss.NewStyle().Foreground(c.theme.commentDescColor).
		Faint(true).Render(desc)

	content, err := c.converter.ConvertString(comment.Text)
	if err != nil {
		// TODO: should log
		content = comment.Text
	}
	content = lipgloss.NewStyle().Border(c.theme.border.style, false, false, true, false).
		Width(c.model.Width()).Foreground(c.theme.commentContentColor).Render(content)

	return fmt.Sprintf("%s\n%s", desc, content)
}
