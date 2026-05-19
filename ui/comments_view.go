package ui

import (
	"context"
	"fmt"
	"html"
	"net/url"
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

type itemLoadSuccessMsg struct {
	item domain.Item
}

type itemLoadFailedMsg struct {
	err error
}

type commentsView struct {
	itemID      int64
	item        domain.Item
	itemLoaded  bool
	itemLoadErr error
	client      *hn.Client
	theme       config.Theme
	hotkey      config.Hotkey
	model       viewport.Model
	converter   *converter.Converter
	msg         commentsMsg
	spinner     spinner.Model
	tree        *commentsTree
}

func newCommentsView(itemID int64, client *hn.Client, theme config.Theme, hotkey config.Hotkey) *commentsView {
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
		itemID:    itemID,
		client:    client,
		theme:     theme,
		hotkey:    hotkey,
		model:     vp,
		converter: converter,
		spinner:   s,
		tree:      newCommentsTree(itemID),
	}
}

func (c *commentsView) Init() tea.Cmd {
	return tea.Batch(
		c.spinner.Tick,
		c.fetchItem(),
	)
}

func (c *commentsView) Update(msg tea.Msg) (*commentsView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case itemLoadSuccessMsg:
		c.item = msg.item
		c.itemLoaded = true
		c.render()
		return c, c.fetchComments()
	case itemLoadFailedMsg:
		c.itemLoadErr = msg.err
		return c, nil
	case spinner.TickMsg:
		if c.hasLoadingComments() || !c.itemLoaded {
			c.spinner, cmd = c.spinner.Update(msg)
			c.render()
		}
		return c, cmd
	case commentsMsg:
		if msg.item.ID != c.itemID {
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
	if c.itemLoadErr != nil {
		return fmt.Sprintf("Load Item Failed: %s", c.itemLoadErr.Error())
	}
	if !c.itemLoaded {
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(c.model.Width()).
			Render(fmt.Sprintf("%s Loading...", c.spinner.View()))
	}

	return c.model.View()
}

func (c *commentsView) setSize(width, height int) {
	c.model.SetHeight(height)
	c.model.SetWidth(width)
	c.render()
	c.ensureSelectedVisible()
}

func (c *commentsView) fetchItem() tea.Cmd {
	return func() tea.Msg {
		item, err := c.client.FetchItem(context.Background(), c.itemID)
		if err != nil {
			return itemLoadFailedMsg{err: err}
		}
		return itemLoadSuccessMsg{item: item}
	}
}

func (c commentsView) fetchComments() tea.Cmd {
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
	if msg.parentID == c.itemID {
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
		c.ensureSelectedVisible()
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
	c.ensureSelectedVisible()
}

func (c *commentsView) itemDomain() string {
	if c.item.URL == "" {
		return ""
	}
	u, err := url.Parse(c.item.URL)
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

func (c *commentsView) renderItemHeader() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Padding(0, 1).
		Foreground(c.theme.Item.TitleSelectedColor).Bold(true)
	descStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(c.theme.Item.DescColor)

	domain := c.itemDomain()
	if domain != "" {
		domain = fmt.Sprintf(" (%s)", domain)
	}

	titleText := fmt.Sprintf("%s%s", c.item.Title, domain)
	// Word wrap title based on viewport width
	wrappedTitle := lipgloss.NewStyle().Width(max(1, c.model.Width()-2)).Render(titleText)
	fmt.Fprintln(&s, titleStyle.Render(wrappedTitle))

	desc := fmt.Sprintf("%d points by %s %s", c.item.Score, c.item.By, c.item.TimeAgo())
	if c.item.Descendants == 1 {
		desc = fmt.Sprintf("%s | 1 comment", desc)
	} else if c.item.Descendants > 1 {
		desc = fmt.Sprintf("%s | %d comments", desc, c.item.Descendants)
	}
	fmt.Fprintln(&s, descStyle.Render(desc))

	// If there is text in the item (e.g. Ask HN post content), convert it and display it!
	if c.item.Text != "" {
		var content string
		var err error
		content, err = c.converter.ConvertString(c.item.Text)
		if err != nil {
			content = c.item.Text
		}
		content = strings.TrimSpace(content)
		content = html.UnescapeString(content)

		if content != "" {
			fmt.Fprintln(&s)
			textStyle := lipgloss.NewStyle().
				PaddingLeft(2).
				Width(max(1, c.model.Width()-4)).
				Foreground(c.theme.Comment.ContentColor)
			fmt.Fprintln(&s, textStyle.Render(content))
		}
	}

	// Add a beautiful separator line
	separator := lipgloss.NewStyle().
		Foreground(c.theme.TitleBar.DivideColor).
		Render(strings.Repeat("─", max(1, c.model.Width())))
	fmt.Fprintln(&s, separator)

	return s.String()
}

func (c *commentsView) render() {
	if !c.itemLoaded {
		return
	}

	var s strings.Builder
	line := 0

	// 1. Render the item info at the top (in the front)!
	headerStr := c.renderItemHeader()
	fmt.Fprint(&s, headerStr)
	line += strings.Count(headerStr, "\n")

	// 2. Render comments or loading comments state
	if c.tree.RootCount() == 0 {
		if c.msg.state == stateLoading {
			c.appendLines(&s, &line, fmt.Sprintf("  %s Loading comments...", c.spinner.View()))
		} else if c.msg.state == stateLoadFailed {
			c.appendLines(&s, &line, fmt.Sprintf("  Load comments failed: %s", c.msg.err.Error()))
		} else if c.msg.state == stateLoadSuccess {
			c.appendLines(&s, &line, "  No comments")
		}
	} else {
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
	}

	c.model.SetContent(strings.TrimRight(s.String(), "\n"))
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
	visibleList := c.tree.Visible()
	for i, visible := range visibleList {
		if visible.id == c.tree.SelectedID() {
			if i == 0 {
				c.model.EnsureVisible(0, 0, 0)
			} else {
				c.model.EnsureVisible(visible.line, 0, 0)
			}
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
		c.ensureSelectedVisible()
		if req.ok {
			return c, c.fetchChildren(req.parentID, req.ids)
		}
		return c, nil
	}

	if key.Matches(msg, c.hotkey.SelectParent) {
		c.tree.SelectParent()
		c.render()
		c.ensureSelectedVisible()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.PrevComment) {
		c.tree.SelectVisible(-1)
		c.render()
		c.ensureSelectedVisible()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.NextComment) {
		c.tree.SelectVisible(1)
		c.render()
		c.ensureSelectedVisible()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.PrevSiblingComment) {
		c.tree.SelectSibling(-1)
		c.render()
		c.ensureSelectedVisible()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.NextSiblingComment) {
		c.tree.SelectSibling(1)
		c.render()
		c.ensureSelectedVisible()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.GoToStart) {
		c.tree.SelectFirst()
		c.render()
		c.ensureSelectedVisible()
		return c, nil
	}

	if key.Matches(msg, c.hotkey.GoToEnd) {
		c.tree.SelectLast()
		c.render()
		c.ensureSelectedVisible()
		return c, nil
	}

	return c, nil
}
