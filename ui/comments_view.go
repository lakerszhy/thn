package ui

import (
	"context"
	"fmt"
	"html"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type commentsView struct {
	item      domain.Item
	client    *hn.Client
	theme     config.Theme
	hotkey    config.Hotkey
	model     viewport.Model
	converter *converter.Converter
	msg       commentsMsg
	spinner   spinner.Model
	tree      *commentsTree
}

func newCommentsView(item domain.Item, client *hn.Client, theme config.Theme, hotkey config.Hotkey) *commentsView {
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
		hotkey:    hotkey,
		model:     vp,
		converter: converter,
		spinner:   s,
		tree:      newCommentsTree(item.ID),
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
		if c.hasLoadingComments() {
			c.spinner, cmd = c.spinner.Update(msg)
			c.render()
		}
		return c, cmd
	case commentsMsg:
		if msg.item.ID != c.item.ID {
			return c, nil
		}

		c.msg = msg
		c.applyCommentsMsg(msg)
		if msg.state == stateLoadingMore {
			return c, c.spinner.Tick
		}
		return c, nil
	case tea.KeyPressMsg:
		c, cmd = c.onKeyPressMsg(msg)
		return c, cmd
	}

	c.model, cmd = c.model.Update(msg)
	return c, cmd
}

func (c *commentsView) View() string {
	if c.msg.state == stateLoading && c.tree.RootCount() == 0 {
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(c.model.Width()).
			Render(fmt.Sprintf("%s Loading...", c.spinner.View()))
	}
	if c.msg.state == stateLoadFailed && c.tree.RootCount() == 0 {
		return fmt.Sprintf("Load Failed: %s", c.msg.err.Error())
	}

	return c.model.View()
}

func (c *commentsView) setSize(width, height int) {
	c.model.SetHeight(height)
	c.model.SetWidth(width)
	c.render()
}

func (c commentsView) fetch() tea.Cmd {
	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newCommentsLoadingMsg(c.item)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := c.client.FetchComments(context.Background(), c.item.KIDs)
		if err != nil {
			return newCommentsLoadFailedMsg(c.item, err)
		}
		return newCommentsLoadSuccessMsg(c.item, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (c commentsView) fetchChildren(parentID int64, ids []int64) tea.Cmd {
	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newCommentChildrenLoadingMsg(c.item, parentID)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := c.client.FetchComments(context.Background(), ids)
		if err != nil {
			return newCommentChildrenLoadFailedMsg(c.item, parentID, err)
		}
		return newCommentChildrenLoadSuccessMsg(c.item, parentID, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (c *commentsView) applyCommentsMsg(msg commentsMsg) {
	if msg.parentID == c.item.ID {
		switch msg.state {
		case stateLoading:
			return
		case stateLoadFailed:
			c.model.SetContent(fmt.Sprintf("Load Failed: %s", msg.err.Error()))
			return
		case stateLoadSuccess:
			c.tree.SetRoots(msg.comments)
		}
		c.render()
		return
	}

	switch msg.state {
	case stateLoadingMore:
		c.tree.StartLoading(msg.parentID)
	case stateLoadMoreSuccess:
		c.tree.SetChildren(msg.parentID, msg.comments)
	case stateLoadMoreFailed:
		c.tree.FailLoading(msg.parentID, msg.err)
	}

	c.render()
}

func (c *commentsView) render() {
	if c.tree.RootCount() == 0 {
		if c.msg.state == stateLoadSuccess {
			c.model.SetContent("No comments")
		}
		return
	}

	var s strings.Builder
	line := 0
	for _, visible := range c.tree.Visible() {
		node := c.tree.Node(visible.id)
		if node == nil {
			continue
		}

		c.tree.SetVisibleLine(visible.id, line)
		c.appendLines(&s, &line, c.renderCommentHeader(node, visible.depth))
		c.appendLines(&s, &line, c.renderCommentBody(node.comment, visible.depth), "")

		if node.expanded {
			if node.loading {
				c.appendLines(&s, &line, fmt.Sprintf("%s%s Loading...", strings.Repeat("  ", visible.depth+1), c.spinner.View()), "")
			}
			if node.err != nil {
				c.appendLines(&s, &line, fmt.Sprintf("%sLoad failed: %s", strings.Repeat("  ", visible.depth+1), node.err.Error()), "")
			}
		}
	}

	c.model.SetContent(strings.TrimRight(s.String(), "\n"))
	c.ensureSelectedVisible()
}

func (c *commentsView) appendLines(s *strings.Builder, line *int, lines ...string) {
	for _, text := range lines {
		fmt.Fprintln(s, text)
		*line += strings.Count(text, "\n") + 1
	}
}

func (c *commentsView) renderCommentHeader(node *commentNode, depth int) string {
	marker := " "
	if len(node.comment.KIDs) > 0 {
		marker = "+"
		if node.expanded {
			marker = "-"
		}
	}

	desc := fmt.Sprintf("%s %s", node.comment.By, node.comment.TimeAgo())
	if len(node.comment.KIDs) > 0 {
		desc = fmt.Sprintf("%s [%d]", desc, len(node.comment.KIDs))
	}

	header := fmt.Sprintf("%s%s %s", strings.Repeat("  ", depth), marker, desc)
	style := lipgloss.NewStyle().Foreground(c.theme.Comment.DescColor)
	if node.comment.ID == c.tree.SelectedID() {
		style = style.Foreground(c.theme.Comment.DescSelectedColor).Bold(true)
	}
	return style.Render(header)
}

func (c *commentsView) renderCommentBody(comment domain.Comment, depth int) string {
	var content string

	switch {
	case comment.Deleted:
		content = "[deleted]"
	case comment.Dead:
		content = "[dead]"
	default:
		var err error
		content, err = c.converter.ConvertString(comment.Text)
		if err != nil {
			content = comment.Text
		}
		content = strings.TrimSpace(content)
		content = html.UnescapeString(content)
	}

	if content == "" {
		content = "[empty]"
	}

	return lipgloss.NewStyle().
		PaddingLeft(depth*2 + 2).
		Width(max(1, c.model.Width()-depth*2-2)).
		Foreground(c.theme.Comment.ContentColor).
		Render(content)
}

func (c *commentsView) ensureSelectedVisible() {
	for _, visible := range c.tree.Visible() {
		if visible.id == c.tree.SelectedID() {
			c.model.EnsureVisible(visible.line, 0, 0)
			return
		}
	}
}

func (c *commentsView) hasLoadingComments() bool {
	return c.msg.state == stateLoading || c.tree.HasLoading()
}

func (c *commentsView) onKeyPressMsg(msg tea.KeyPressMsg) (*commentsView, tea.Cmd) {
	if key.Matches(msg, c.hotkey.ToggleSubComments) {
		req := c.tree.ToggleSelected()
		c.render()
		if req.ok {
			return c, c.fetchChildren(req.parentID, req.ids)
		}
		return c, nil
	}

	if key.Matches(msg, c.hotkey.SelectParent) {
		c.tree.SelectParent()
		c.render()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.PrevComment) {
		c.tree.SelectVisible(-1)
		c.render()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.NextComment) {
		c.tree.SelectVisible(1)
		c.render()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.PrevSiblingComment) {
		c.tree.SelectSibling(-1)
		c.render()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.NextSiblingComment) {
		c.tree.SelectSibling(1)
		c.render()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.GoToStart) {
		c.tree.SelectFirst()
		c.render()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.GoToEnd) {
		c.tree.SelectLast()
		c.render()
		return c, nil
	}

	return c, nil
}
