package comments

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pkg/browser"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type View struct {
	itemID int64
	client *hn.Client
	theme  config.Theme
	hotkey config.Hotkey

	model   viewport.Model
	spinner spinner.Model

	msg     commentsMsg
	itemMsg itemMsg
	tree    *tree
}

func NewView(itemID int64, client *hn.Client, theme config.Theme, hotkey config.Hotkey) *View {
	vp := viewport.New()
	vp.SoftWrap = true

	s := spinner.New()
	s.Spinner = spinner.Dot

	return &View{
		itemID:  itemID,
		client:  client,
		theme:   theme,
		hotkey:  hotkey,
		model:   vp,
		spinner: s,
		tree:    newTree(itemID),
	}
}

func (c *View) Init() tea.Cmd {
	return tea.Batch(
		c.spinner.Tick,
		c.fetchItem(),
	)
}

func (c *View) Update(msg tea.Msg) (*View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case itemMsg:
		if msg.itemID != c.itemID {
			return c, nil
		}

		c.itemMsg = msg
		if msg.state == stateLoadSuccess {
			c.render()
			return c, c.fetchComments()
		}
		return c, nil
	case spinner.TickMsg:
		if c.msg.state == stateLoading || c.itemMsg.state == stateLoading {
			c.spinner, cmd = c.spinner.Update(msg)
			c.render()
		}
		return c, cmd
	case commentsMsg:
		return c.onCommentsMsg(msg)
	case tea.KeyPressMsg:
		c, cmd = c.onKeyPressMsg(msg)
		return c, cmd
	}

	c.model, cmd = c.model.Update(msg)
	return c, cmd
}

func (c *View) View() string {
	if c.itemMsg.state == stateLoadFailed {
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(c.model.Width()).
			Render(fmt.Sprintf("Load Item Failed: %s", c.itemMsg.err.Error()))
	}

	if c.itemMsg.state != stateLoadSuccess {
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(c.model.Width()).
			Render(fmt.Sprintf("%s Loading...", c.spinner.View()))
	}

	return c.model.View()
}

func (c *View) SetSize(width, height int) {
	c.model.SetHeight(height)
	c.model.SetWidth(width)
	c.render()
	c.ensureSelectedVisible()
}

func (c *View) fetchItem() tea.Cmd {
	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newItemLoadingMsg(c.itemID)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		item, err := c.client.FetchItem(context.Background(), c.itemID)
		if err != nil {
			return newItemLoadFailedMsg(c.itemID, err)
		}
		return newItemLoadSuccessMsg(c.itemID, item)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (c *View) fetchComments() tea.Cmd {
	if c.itemMsg.state != stateLoadSuccess {
		return nil
	}

	if !c.itemMsg.item.HasComments() {
		return nil
	}

	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newCommentsLoadingMsg(c.itemID)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := c.client.FetchComments(context.Background(), c.itemMsg.item.KIDs)
		if err != nil {
			return newCommentsLoadFailedMsg(c.itemID, err)
		}
		return newCommentsLoadSuccessMsg(c.itemID, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (c *View) fetchSubComments(comment domain.Comment) tea.Cmd {
	if c.itemMsg.state != stateLoadSuccess {
		return nil
	}

	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newSubCommentsLoadingMsg(c.itemID, comment.ID)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := c.client.FetchComments(context.Background(), comment.KIDs)
		if err != nil {
			return newSubCommentsLoadFailedMsg(c.itemID, comment.ID, err)
		}
		return newSubCommentsLoadSuccessMsg(c.itemID, comment.ID, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (c *View) onCommentsMsg(msg commentsMsg) (*View, tea.Cmd) {
	if msg.itemID != c.itemID {
		return c, nil
	}

	c.msg = msg

	// comments of item
	if msg.isRoot() {
		switch msg.state {
		case stateLoadFailed:
			c.model.SetContent(fmt.Sprintf("Load Failed: %s", msg.err.Error()))
			return c, nil
		case stateLoadSuccess:
			c.tree.SetRoots(msg.comments)
			c.render()
			c.ensureSelectedVisible()
			return c, nil
		}

		return c, nil
	}

	var cmd tea.Cmd

	// sub comments of some comment
	switch msg.state {
	case stateLoading:
		c.tree.StartLoading(msg.commentID)
		cmd = c.spinner.Tick
	case stateLoadSuccess:
		c.tree.SetChildren(msg.commentID, msg.comments)
	case stateLoadFailed:
		c.tree.FailLoading(msg.commentID, msg.err)
	}

	c.render()
	c.ensureSelectedVisible()

	return c, cmd
}

func (c *View) renderItem() string {
	if c.itemMsg.state != stateLoadSuccess {
		return ""
	}

	var s strings.Builder

	domain := c.itemMsg.item.Domain()
	if domain != "" {
		domain = fmt.Sprintf(" (%s)", domain)
	}

	titleText := fmt.Sprintf("%s%s", c.itemMsg.item.Title, domain)
	titleStyle := lipgloss.NewStyle().Padding(0, 1).
		Width(max(1, c.model.Width())).Bold(true).
		Foreground(c.theme.Item.TitleSelectedColor)
	fmt.Fprintln(&s, titleStyle.Render(titleText))

	desc := c.itemMsg.item.Description()
	descStyle := lipgloss.NewStyle().Padding(0, 1).
		Width(max(1, c.model.Width())).
		Foreground(c.theme.Item.DescColor)
	fmt.Fprintln(&s, descStyle.Render(desc))

	if c.itemMsg.item.Text != "" {
		fmt.Fprintln(&s)
		textStyle := lipgloss.NewStyle().
			PaddingLeft(2).
			Width(max(1, c.model.Width()-4)).
			Foreground(c.theme.Comment.ContentColor)
		fmt.Fprintln(&s, textStyle.Render(c.itemMsg.item.Text))
	}

	separator := lipgloss.NewStyle().
		Foreground(c.theme.TitleBar.DivideColor).
		Render(strings.Repeat("─", max(1, c.model.Width())))
	fmt.Fprintln(&s, separator)

	return s.String()
}

func (c *View) render() {
	if c.itemMsg.state != stateLoadSuccess {
		return
	}

	var s strings.Builder
	line := 0

	// 1. Render the item info at the top (in the front)!
	headerStr := c.renderItem()
	fmt.Fprint(&s, headerStr)
	line += strings.Count(headerStr, "\n")

	// 2. Render comments or loading comments state
	if c.tree.RootCount() == 0 {
		if !c.itemMsg.item.HasComments() {
			c.appendLines(&s, &line, "  No comments")
		} else {
			switch c.msg.state {
			case stateLoading:
				c.appendLines(&s, &line, fmt.Sprintf("  %s Loading comments...", c.spinner.View()))
			case stateLoadFailed:
				c.appendLines(&s, &line, fmt.Sprintf("  Load comments failed: %s", c.msg.err.Error()))
			case stateLoadSuccess:
				c.appendLines(&s, &line, "  No comments")
			}
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

func (c *View) appendLines(s *strings.Builder, line *int, lines ...string) {
	for _, text := range lines {
		fmt.Fprintln(s, text)
		*line += strings.Count(text, "\n") + 1
	}
}

func (c *View) renderCommentHeader(node *node, depth int) string {
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

func (c *View) renderCommentBody(comment domain.Comment, depth int) string {
	var content string

	switch {
	case comment.Deleted:
		content = "[deleted]"
	case comment.Dead:
		content = "[dead]"
	default:
		content = comment.Text
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

func (c *View) ensureSelectedVisible() {
	for i, visible := range c.tree.Visible() {
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

func (c *View) onKeyPressMsg(msg tea.KeyPressMsg) (*View, tea.Cmd) {
	if c.itemMsg.state != stateLoadSuccess {
		return c, nil
	}

	var cmd tea.Cmd
	reRender := true

	switch {
	case key.Matches(msg, c.hotkey.ToggleSubComments):
		comment := c.tree.ToggleSelected()
		if comment != nil {
			cmd = c.fetchSubComments(*comment)
		}

	case key.Matches(msg, c.hotkey.SelectParent):
		c.tree.SelectParent()

	case key.Matches(msg, c.hotkey.PrevComment):
		c.tree.SelectVisible(-1)

	case key.Matches(msg, c.hotkey.NextComment):
		c.tree.SelectVisible(1)

	case key.Matches(msg, c.hotkey.PrevSiblingComment):
		c.tree.SelectSibling(-1)

	case key.Matches(msg, c.hotkey.NextSiblingComment):
		c.tree.SelectSibling(1)

	case key.Matches(msg, c.hotkey.GoToStart):
		c.tree.SelectFirst()

	case key.Matches(msg, c.hotkey.GoToEnd):
		c.tree.SelectLast()

	case key.Matches(msg, c.hotkey.OpenHNInBrowser):
		reRender = false
		c.openURL(c.itemMsg.item.URLInHN())

	case key.Matches(msg, c.hotkey.OpenDomainInBrowser):
		reRender = false
		if c.itemMsg.item.URL != "" {
			c.openURL(c.itemMsg.item.URL)
		}

	case key.Matches(msg, c.hotkey.OpenCommentInBrowser):
		reRender = false
		node := c.tree.Node(c.tree.SelectedID())
		if node != nil {
			c.openURL(node.comment.URLInHN(c.itemMsg.item.ID))
		}
	}

	if reRender {
		c.render()
		c.ensureSelectedVisible()
	}

	return c, cmd
}

func (c *View) openURL(u string) {
	if err := browser.OpenURL(u); err != nil {
		slog.Error("open url failed", "url", u, "error", err)
	}
}
