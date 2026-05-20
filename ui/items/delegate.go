package items

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
)

type listItem struct {
	domain.Item
}

func (listItem) FilterValue() string {
	return ""
}

type loadMoreItem struct {
}

func (l loadMoreItem) FilterValue() string {
	return ""
}

type delegate struct {
	normalTitle    lipgloss.Style
	selectedTitle  lipgloss.Style
	normalDesc     lipgloss.Style
	selectedDesc   lipgloss.Style
	normalDomain   lipgloss.Style
	selectedDomain lipgloss.Style

	ellipsis string

	msg itemsMsg
}

func newDeletage(t config.Theme) *delegate {
	//nolint:mnd // 6: 1 for ">", 3 for index, 1 for ".", 1 for space
	desc := lipgloss.NewStyle().PaddingLeft(6)
	return &delegate{
		// 1 for ">"
		normalTitle:    lipgloss.NewStyle().PaddingLeft(1).Foreground(t.Item.TitleColor),
		normalDesc:     desc.Foreground(t.Item.DescColor),
		selectedTitle:  lipgloss.NewStyle().Foreground(t.Item.TitleSelectedColor),
		selectedDesc:   desc.Foreground(t.Item.DescSelectedColor),
		normalDomain:   desc.Foreground(t.Item.DomainColor),
		selectedDomain: desc.Foreground(t.Item.DomainSelectedColor),
		ellipsis:       "...",
	}
}

func (d *delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if m.Width() <= 0 {
		return
	}

	var selected = index == m.Index()

	if _, ok := item.(loadMoreItem); ok {
		fmt.Fprintf(w, "%s", d.renderLoadMore(selected))
		return
	}

	var title, desc, domain string

	if i, ok := item.(listItem); ok {
		title = fmt.Sprintf("%3d. %s", index+1, i.Title)
		domain = i.Domain()
		desc = i.Description()
	} else {
		return
	}

	if domain != "" {
		domain = fmt.Sprintf(" (%s)", domain)
	}

	if selected {
		title = d.selectedTitle.Render(">" + title)
		domain = d.selectedDomain.UnsetPadding().Render(domain)
		desc = d.selectedDesc.Render(desc)
	} else {
		title = d.normalTitle.Render(title)
		domain = d.normalDomain.UnsetPadding().Render(domain)
		desc = d.normalDesc.Render(desc)
	}

	textwidth := m.Width() - d.normalTitle.GetHorizontalPadding()

	title = fmt.Sprintf("%s%s", title, domain)
	title = ansi.Truncate(title, textwidth, d.ellipsis)

	desc = ansi.Truncate(desc, textwidth, d.ellipsis)

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func (d *delegate) Height() int {
	return 2 //nolint:mnd // height of item
}

func (d *delegate) Spacing() int {
	return 0
}

func (d *delegate) Update(msg tea.Msg, _ *list.Model) tea.Cmd {
	if msg, ok := msg.(itemsMsg); ok {
		d.msg = msg
	}

	return nil
}

func (d *delegate) renderLoadMore(selected bool) string {
	content := "More"
	switch d.msg.state {
	case stateLoadingMore:
		content = "Loading..."
	case stateLoadMoreFailed:
		content = fmt.Sprintf("Load more failed: %s", d.msg.err.Error())
	}

	content = lipgloss.NewStyle().PaddingLeft(1).Render(content)
	if selected {
		content = d.selectedTitle.Render(">" + content)
	} else {
		content = d.normalTitle.Render(content)
	}

	return content
}
