package ui

import "github.com/lakerszhy/thn/domain"

const (
	stateLoading = iota
	stateLoadSuccess
	stateLoadFailed
)

type itemsMsg struct {
	category domain.Category
	items    []domain.Item
	baseMsg
}

func newItemsLoadingMsg(cat domain.Category) itemsMsg {
	return itemsMsg{
		baseMsg:  baseMsg{state: stateLoading},
		category: cat,
	}
}

func newItemsLoadSuccessMsg(cat domain.Category, items []domain.Item) itemsMsg {
	return itemsMsg{
		baseMsg:  baseMsg{state: stateLoadSuccess},
		category: cat,
		items:    items,
	}
}

func newItemsLoadFailedMsg(cat domain.Category, err error) itemsMsg {
	return itemsMsg{
		baseMsg:  baseMsg{state: stateLoadFailed, err: err},
		category: cat,
	}
}

type itemSelectedMsg domain.Item

type commentsMsg struct {
	item     domain.Item
	comments []domain.Item
}

type state int

type baseMsg struct {
	state state
	err   error
}

func (b baseMsg) IsLoading() bool {
	return b.state == stateLoading
}
