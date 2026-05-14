package ui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
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
	msg       commentsMsg
	spinner   spinner.Model
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

	s := spinner.New()
	s.Spinner = spinner.Dot

	return &commentsView{
		item:      item,
		client:    client,
		theme:     theme,
		model:     vp,
		converter: converter,
		spinner:   s,
	}
}

func (c *commentsView) Init() tea.Cmd {
	return tea.Batch(
		c.spinner.Tick,
		c.fetch(),
	)
}

func (c *commentsView) Update(msg tea.Msg) (*commentsView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if c.msg.isLoading() {
			c.spinner, cmd = c.spinner.Update(msg)
		}
		return c, cmd
	case commentsMsg:
		if msg.item.ID != c.item.ID {
			return c, nil
		}

		c.msg = msg

		if msg.isSuccess() {
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

	c.model, cmd = c.model.Update(msg)
	return c, cmd
}

func (c *commentsView) View() string {
	switch c.msg.state {
	case stateLoading:
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(c.model.Width()).
			Render(fmt.Sprintf("%s Loading...", c.spinner.View()))
	case stateLoadFailed:
		return fmt.Sprintf("Load Failed: %s", c.msg.err.Error())
	case stateLoadSuccess:
		return c.model.View()
	}

	return c.model.View()
}

func (c *commentsView) setSize(width, height int) {
	c.model.SetHeight(height)
	c.model.SetWidth(width)
}

func (c commentsView) fetch() tea.Cmd {
	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newCommentsLoadingMsg(c.item)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := c.client.FetchComments(context.Background(), c.item)
		if err != nil {
			return newCommentsLoadFailedMsg(c.item, err)
		}
		return newCommentsLoadSuccessMsg(c.item, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
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
