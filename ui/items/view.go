package items

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pkg/browser"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type View struct {
	theme    config.Theme
	hotkey   config.Hotkey
	category domain.Category
	client   *hn.Client
	msg      itemsMsg
	spinner  spinner.Model

	model list.Model
}

func NewView(category domain.Category, client *hn.Client, theme config.Theme, hotkey config.Hotkey) *View {
	model := list.New(nil, newDeletage(theme), 0, 0)
	model.SetShowTitle(false)
	model.SetFilteringEnabled(false)
	model.SetShowStatusBar(false)
	model.SetShowPagination(false)
	model.SetShowHelp(false)
	model.DisableQuitKeybindings()
	model.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys(hotkey.PrevItem...)),
		CursorDown: key.NewBinding(key.WithKeys(hotkey.NextItem...)),
		PrevPage:   key.NewBinding(key.WithKeys(hotkey.PrevPage...)),
		NextPage:   key.NewBinding(key.WithKeys(hotkey.NextPage...)),
		GoToStart:  key.NewBinding(key.WithKeys(hotkey.GoToStart...)),
		GoToEnd:    key.NewBinding(key.WithKeys(hotkey.GoToEnd...)),
	}

	s := spinner.New()
	s.Spinner = spinner.Dot

	return &View{
		category: category,
		hotkey:   hotkey,
		client:   client,
		theme:    theme,
		model:    model,
		spinner:  s,
	}
}

func (t *View) Init() tea.Cmd {
	return tea.Batch(
		t.spinner.Tick,
		t.fetch(),
	)
}

func (t *View) Update(msg tea.Msg) (*View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if t.msg.state == stateLoading || t.msg.state == stateLoadingMore {
			t.spinner, cmd = t.spinner.Update(msg)
		}
		return t, cmd
	case itemsMsg:
		return t.onItemsMsg(msg)
	case tea.KeyPressMsg:
		return t.onKeyPressMsg(msg)
	}

	return t, nil
}

func (t *View) View() string {
	switch t.msg.state {
	case stateLoading:
		return lipgloss.NewStyle().Align(lipgloss.Center).Width(t.model.Width()).
			Render(fmt.Sprintf("%s Loading...", t.spinner.View()))
	case stateLoadFailed:
		return fmt.Sprintf("Load Failed: %s", t.msg.err.Error())
	}
	return t.model.View()
}

func (t *View) SetSize(width int, height int) {
	t.model.SetWidth(width)
	t.model.SetHeight(height)
}

func (t *View) onItemsMsg(msg itemsMsg) (*View, tea.Cmd) {
	if msg.category != t.category {
		return t, nil
	}

	t.msg = msg

	switch msg.state {
	case stateLoading, stateLoadFailed:
		t.model.SetItems(nil)
	case stateLoadSuccess, stateLoadMoreSuccess:
		t.onItemsLoaded(msg.pagedItems)
	}

	// handle by delegate
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t *View) onItemsLoaded(pagedItems domain.PagedItems) {
	items := make([]list.Item, 0, len(t.model.Items())+len(pagedItems.Items))

	for _, v := range t.model.Items() {
		if _, ok := v.(listItem); ok {
			items = append(items, v)
		}
	}

	for _, v := range pagedItems.Items {
		items = append(items, listItem{Item: v})
	}

	if pagedItems.HasMore() {
		items = append(items, loadMoreItem{})
	}

	t.model.SetItems(items)
}

func (t *View) onKeyPressMsg(msg tea.KeyPressMsg) (*View, tea.Cmd) {
	key := msg.String()

	if slices.Contains(t.hotkey.OpenComments, key) {
		switch i := t.model.SelectedItem().(type) {
		case listItem:
			return t, func() tea.Msg {
				return NewItemSelectedMsg(i.Item, false)
			}
		case loadMoreItem:
			return t, t.fetchMore()
		}
	}

	if slices.Contains(t.hotkey.OpenCommentsFullscreen, key) {
		if i, ok := t.model.SelectedItem().(listItem); ok {
			return t, func() tea.Msg {
				return NewItemSelectedMsg(i.Item, true)
			}
		}
		return t, nil
	}

	if slices.Contains(t.hotkey.OpenHNInBrowser, key) {
		var item listItem
		var ok bool

		if item, ok = t.model.SelectedItem().(listItem); !ok {
			return t, nil
		}

		if err := browser.OpenURL(item.URLInHN()); err != nil {
			slog.Error("open item url failed", "id", item.ID, "url", item.URLInHN(), "error", err)
		}

		return t, nil
	}

	if slices.Contains(t.hotkey.OpenDomainInBrowser, key) {
		var item listItem
		var ok bool

		if item, ok = t.model.SelectedItem().(listItem); !ok {
			return t, nil
		}

		if item.URL == "" {
			return t, nil
		}

		if err := browser.OpenURL(item.URL); err != nil {
			slog.Error("open item url failed", "id", item.ID, "url", item.URL, "error", err)
		}

		return t, nil
	}

	if slices.Contains(t.hotkey.Refresh, key) && t.msg.state != stateLoading {
		return t, tea.Batch(t.spinner.Tick, t.fetch())
	}

	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t *View) fetch() tea.Cmd {
	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newLoadingMsg(t.category)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := t.client.FetchItems(context.Background(), t.category, domain.NewPagination())
		if err != nil {
			return newLoadFailedMsg(t.category, err)
		}
		return newLoadSuccessMsg(t.category, items)
	}
	cmds = append(cmds, cmd)

	// If use tea.Batch, the spinner.Tick may not be executed.
	return tea.Sequence(cmds...)
}

func (t *View) fetchMore() tea.Cmd {
	if t.msg.state == stateLoading || t.msg.state == stateLoadingMore {
		return nil
	}

	pagination := t.msg.pagedItems.Pagination
	if !pagination.HasMore() {
		return nil
	}

	pagination = pagination.Next()

	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newLoadingMoreMsg(t.category)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := t.client.FetchItems(context.Background(), t.category, pagination)
		if err != nil {
			return newLoadMoreFailedMsg(t.category, err)
		}
		return newLoadMoreSuccessMsg(t.category, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}
