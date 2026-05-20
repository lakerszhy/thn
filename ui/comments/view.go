package comments

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

type View struct {
	itemID int64
	client *hn.Client
	theme  config.Theme
	hotkey config.Hotkey

	model     viewport.Model
	spinner   spinner.Model
	converter *converter.Converter

	msg     commentsMsg
	itemMsg itemMsg
	tree    *tree
}

func NewView(itemID int64, client *hn.Client, theme config.Theme, hotkey config.Hotkey) *View {
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

	return &View{
		itemID:    itemID,
		client:    client,
		theme:     theme,
		hotkey:    hotkey,
		model:     vp,
		converter: converter,
		spinner:   s,
		tree:      newTree(itemID),
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
		if c.hasLoadingComments() {
			c.spinner, cmd = c.spinner.Update(msg)
			c.render()
		}
		return c, cmd
	case commentsMsg:
		if msg.itemID != c.itemID {
			return c, nil
		}

		c.msg = msg
		c.applyCommentsMsg(msg)
		if msg.state == stateLoading {
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

func (c *View) fetchSubComments(parentID int64, ids []int64) tea.Cmd {
	if c.itemMsg.state != stateLoadSuccess {
		return nil
	}

	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newSubCommentsLoadingMsg(c.itemID, parentID)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := c.client.FetchComments(context.Background(), ids)
		if err != nil {
			return newSubCommentsLoadFailedMsg(c.itemID, parentID, err)
		}
		return newSubCommentsLoadSuccessMsg(c.itemID, parentID, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (c *View) applyCommentsMsg(msg commentsMsg) {
	// comments of item
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

	// sub comments of some comment
	switch msg.state {
	case stateLoading:
		c.tree.StartLoading(msg.parentID)
	case stateLoadSuccess:
		c.tree.SetChildren(msg.parentID, msg.comments)
	case stateLoadFailed:
		c.tree.FailLoading(msg.parentID, msg.err)
	}

	c.render()
	c.ensureSelectedVisible()
}

func (c *View) renderItemHeader() string {
	if c.itemMsg.state != stateLoadSuccess {
		return ""
	}

	var s strings.Builder

	titleStyle := lipgloss.NewStyle().Padding(0, 1).
		Foreground(c.theme.Item.TitleSelectedColor).Bold(true)
	descStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(c.theme.Item.DescColor)

	domain := c.itemMsg.item.Domain()
	if domain != "" {
		domain = fmt.Sprintf(" (%s)", domain)
	}

	titleText := fmt.Sprintf("%s%s", c.itemMsg.item.Title, domain)
	// Word wrap title based on viewport width
	wrappedTitle := lipgloss.NewStyle().Width(max(1, c.model.Width()-2)).Render(titleText)
	fmt.Fprintln(&s, titleStyle.Render(wrappedTitle))

	desc := c.itemMsg.item.Description()
	if c.itemMsg.item.Descendants == 1 {
		desc = fmt.Sprintf("%s | 1 comment", desc)
	} else if c.itemMsg.item.Descendants > 1 {
		desc = fmt.Sprintf("%s | %d comments", desc, c.itemMsg.item.Descendants)
	}
	fmt.Fprintln(&s, descStyle.Render(desc))

	// If there is text in the item (e.g. Ask HN post content), convert it and display it!
	if c.itemMsg.item.Text != "" {
		var content string
		var err error
		content, err = c.converter.ConvertString(c.itemMsg.item.Text)
		if err != nil {
			content = c.itemMsg.item.Text
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

func (c *View) render() {
	if c.itemMsg.state != stateLoadSuccess {
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
		switch c.msg.state {
		case stateLoading:
			c.appendLines(&s, &line, fmt.Sprintf("  %s Loading comments...", c.spinner.View()))
		case stateLoadFailed:
			c.appendLines(&s, &line, fmt.Sprintf("  Load comments failed: %s", c.msg.err.Error()))
		case stateLoadSuccess:
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

func (c *View) ensureSelectedVisible() {
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

func (c *View) hasLoadingComments() bool {
	return c.msg.state == stateLoading || c.tree.HasLoading() || c.itemMsg.state == stateLoading
}

func (c *View) onKeyPressMsg(msg tea.KeyPressMsg) (*View, tea.Cmd) {
	if key.Matches(msg, c.hotkey.ToggleSubComments) {
		req := c.tree.ToggleSelected()
		c.render()
		c.ensureSelectedVisible()
		if req.ok {
			return c, c.fetchSubComments(req.parentID, req.ids)
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
