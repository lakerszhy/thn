package items

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
	"github.com/lakerszhy/thn/hn"
)

type View struct {
	theme      config.Theme
	hotkey     config.Hotkey
	category   domain.Category
	pagination domain.Pagination
	client     *hn.Client
	msg        itemsMsg
	spinner    spinner.Model

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
		CursorUp:   hotkey.PrevItem,
		CursorDown: hotkey.NextItem,
		PrevPage:   hotkey.PrevPage,
		NextPage:   hotkey.NextPage,
		GoToStart:  hotkey.GoToStart,
		GoToEnd:    hotkey.GoToEnd,
	}

	s := spinner.New()
	s.Spinner = spinner.Dot

	return &View{
		category:   category,
		hotkey:     hotkey,
		client:     client,
		pagination: domain.NewPagination(),
		theme:      theme,
		model:      model,
		spinner:    s,
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
		if msg.category != t.category {
			return t, nil
		}

		t.msg = msg

		switch msg.state {
		case stateLoading, stateLoadFailed:
			t.model.SetItems(nil)
			return t, nil
		case stateLoadSuccess:
			items := make([]list.Item, len(msg.items)+1)
			for i, v := range msg.items {
				items[i] = listItem{Item: v}
			}
			items[len(items)-1] = loadMoreItem{}
			t.model.SetItems(items)
		case stateLoadMoreSuccess:
			t.pagination = t.pagination.Next()

			items := make([]list.Item, 0, len(msg.items)+len(t.msg.items))
			items = append(items, t.model.Items()[0:len(t.model.Items())-1]...)
			for _, v := range msg.items {
				items = append(items, listItem{Item: v})
			}
			items = append(items, loadMoreItem{})
			t.model.SetItems(items)
		}
		// handle by delegate, so not return
	case tea.KeyPressMsg:
		if key.Matches(msg, t.hotkey.OpenComments) {
			index := t.model.Index()
			if index < 0 || index >= len(t.model.Items()) {
				return t, nil
			}

			switch i := t.model.Items()[index].(type) {
			case listItem:
				return t, func() tea.Msg {
					return ItemSelectedMsg(i.Item)
				}
			case loadMoreItem:
				return t, t.fetchMore()
			}
		}
	}

	t.model, cmd = t.model.Update(msg)
	return t, cmd
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

func (t *View) fetch() tea.Cmd {
	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newLoadingMsg(t.category)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := t.client.FetchItems(context.Background(), t.category, t.pagination)
		if err != nil {
			return newLoadFailedMsg(t.category, err)
		}
		return newLoadSuccessMsg(t.category, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (t View) fetchMore() tea.Cmd {
	if t.msg.state == stateLoading || t.msg.state == stateLoadingMore {
		return nil
	}

	var cmds []tea.Cmd

	cmd := func() tea.Msg {
		return newLoadingMoreMsg(t.category)
	}
	cmds = append(cmds, cmd)

	cmd = func() tea.Msg {
		items, err := t.client.FetchItems(context.Background(), t.category, t.pagination.Next())
		if err != nil {
			return newLoadMoreFailedMsg(t.category, err)
		}
		return newLoadMoreSuccessMsg(t.category, items)
	}
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}
