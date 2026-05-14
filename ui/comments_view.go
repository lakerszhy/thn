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

	roots      []int64
	nodes      map[int64]*commentNode
	selectedID int64
	visible    []visibleComment
}

type commentNode struct {
	comment  domain.Comment
	parentID int64
	children []int64
	loaded   bool
	loading  bool
	expanded bool
	err      error
}

type visibleComment struct {
	id    int64
	depth int
	line  int
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
		nodes:     make(map[int64]*commentNode),
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
		return c, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", " ", "space", "right", "l":
			return c, c.toggleSelected()
		case "left", "h", "u":
			c.selectParent()
			return c, nil
		case "up", "k":
			c.selectVisible(-1)
			return c, nil
		case "down", "j":
			c.selectVisible(1)
			return c, nil
		case "shift+up", "K", "p":
			c.selectSibling(-1)
			return c, nil
		case "shift+down", "J", "n":
			c.selectSibling(1)
			return c, nil
		}
	}

	c.model, cmd = c.model.Update(msg)
	return c, cmd
}

func (c *commentsView) View() string {
	if c.msg.state == stateLoading && len(c.roots) == 0 {
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(c.model.Width()).
			Render(fmt.Sprintf("%s Loading...", c.spinner.View()))
	}
	if c.msg.state == stateLoadFailed && len(c.roots) == 0 {
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
		items, err := c.client.FetchComments(context.Background(), c.item)
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
		items, err := c.client.FetchCommentsByIDs(context.Background(), ids)
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
			c.roots = c.roots[:0]
			for _, comment := range msg.comments {
				c.upsertComment(comment, c.item.ID)
				c.roots = append(c.roots, comment.ID)
			}
			if c.selectedID == 0 && len(c.roots) > 0 {
				c.selectedID = c.roots[0]
			}
		}
		c.render()
		return
	}

	node, ok := c.nodes[msg.parentID]
	if !ok {
		return
	}

	switch msg.state {
	case stateLoadingMore:
		node.loading = true
		node.err = nil
	case stateLoadMoreSuccess:
		node.children = node.children[:0]
		node.loaded = true
		node.loading = false
		node.err = nil
		for _, comment := range msg.comments {
			c.upsertComment(comment, node.comment.ID)
			node.children = append(node.children, comment.ID)
		}
	case stateLoadMoreFailed:
		node.loading = false
		node.err = msg.err
	}

	c.render()
}

func (c *commentsView) upsertComment(comment domain.Comment, parentID int64) {
	node, ok := c.nodes[comment.ID]
	if !ok {
		c.nodes[comment.ID] = &commentNode{
			comment:  comment,
			parentID: parentID,
		}
		return
	}
	node.comment = comment
	node.parentID = parentID
}

func (c *commentsView) toggleSelected() tea.Cmd {
	node := c.nodes[c.selectedID]
	if node == nil || len(node.comment.KIDs) == 0 {
		return nil
	}

	if node.expanded {
		node.expanded = false
		c.render()
		return nil
	}

	node.expanded = true
	if node.loaded {
		c.render()
		return nil
	}
	if node.loading {
		c.render()
		return nil
	}

	return c.fetchChildren(node.comment.ID, node.comment.KIDs)
}

func (c *commentsView) selectVisible(delta int) {
	index := c.visibleIndex(c.selectedID)
	if index == -1 {
		return
	}

	index += delta
	if index < 0 || index >= len(c.visible) {
		return
	}

	c.selectedID = c.visible[index].id
	c.render()
}

func (c *commentsView) selectParent() {
	node := c.nodes[c.selectedID]
	if node == nil || node.parentID == c.item.ID {
		return
	}

	c.selectedID = node.parentID
	c.render()
}

func (c *commentsView) selectSibling(delta int) {
	node := c.nodes[c.selectedID]
	if node == nil {
		return
	}

	siblings := c.roots
	if parent := c.nodes[node.parentID]; parent != nil {
		siblings = parent.children
	}

	for i, id := range siblings {
		if id != c.selectedID {
			continue
		}
		next := i + delta
		if next < 0 || next >= len(siblings) {
			return
		}
		c.selectedID = siblings[next]
		c.render()
		return
	}
}

func (c *commentsView) render() {
	if len(c.roots) == 0 {
		c.visible = nil
		if c.msg.state == stateLoadSuccess {
			c.model.SetContent("No comments")
		}
		return
	}

	var s strings.Builder
	c.visible = c.visible[:0]
	line := 0
	for _, id := range c.roots {
		c.renderNode(&s, id, 0, &line)
	}

	c.model.SetContent(strings.TrimRight(s.String(), "\n"))
	c.ensureSelectedVisible()
}

func (c *commentsView) renderNode(s *strings.Builder, id int64, depth int, line *int) {
	node := c.nodes[id]
	if node == nil {
		return
	}

	c.visible = append(c.visible, visibleComment{id: id, depth: depth, line: *line})
	c.appendLines(s, line, c.renderCommentHeader(node, depth))
	c.appendLines(s, line, c.renderCommentBody(node.comment, depth), "")

	if node.expanded {
		if node.loading {
			c.appendLines(s, line, fmt.Sprintf("%s%s Loading...", strings.Repeat("  ", depth+1), c.spinner.View()), "")
		}
		if node.err != nil {
			c.appendLines(s, line, fmt.Sprintf("%sLoad failed: %s", strings.Repeat("  ", depth+1), node.err.Error()), "")
		}
		for _, childID := range node.children {
			c.renderNode(s, childID, depth+1, line)
		}
	}
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
	if node.comment.Deleted {
		desc = "deleted"
	}
	if node.comment.Dead {
		desc = "dead"
	}
	if len(node.comment.KIDs) > 0 {
		desc = fmt.Sprintf("%s [%d]", desc, len(node.comment.KIDs))
	}

	header := fmt.Sprintf("%s%s %s", strings.Repeat("  ", depth), marker, desc)
	style := lipgloss.NewStyle().Foreground(c.theme.commentDescColor).Faint(true)
	if node.comment.ID == c.selectedID {
		style = style.Foreground(c.theme.itemTitleSelectedColor).Faint(false).Bold(true)
	}
	return style.Render(header)
}

func (c *commentsView) renderCommentBody(comment domain.Comment, depth int) string {
	content := "[deleted]"
	if !comment.Deleted && !comment.Dead {
		var err error
		content, err = c.converter.ConvertString(comment.Text)
		if err != nil {
			content = comment.Text
		}
		content = strings.TrimSpace(content)
	}
	if content == "" {
		content = "[empty]"
	}

	return lipgloss.NewStyle().
		PaddingLeft(depth*2 + 2).
		Width(max(1, c.model.Width()-depth*2-2)).
		Foreground(c.theme.commentContentColor).
		Render(content)
}

func (c *commentsView) visibleIndex(id int64) int {
	for i, visible := range c.visible {
		if visible.id == id {
			return i
		}
	}
	return -1
}

func (c *commentsView) ensureSelectedVisible() {
	for _, visible := range c.visible {
		if visible.id == c.selectedID {
			c.model.EnsureVisible(visible.line, 0, 0)
			return
		}
	}
}

func (c *commentsView) hasLoadingComments() bool {
	if c.msg.state == stateLoading {
		return true
	}
	for _, node := range c.nodes {
		if node.loading {
			return true
		}
	}
	return false
}
