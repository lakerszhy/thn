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
	baseMsg
}

func newCommentsLoadingMsg(item domain.Item) commentsMsg {
	return commentsMsg{
		baseMsg: baseMsg{state: stateLoading},
		item:    item,
	}
}

func newCommentsLoadSuccessMsg(item domain.Item, comments []domain.Item) commentsMsg {
	return commentsMsg{
		baseMsg:  baseMsg{state: stateLoadSuccess},
		item:     item,
		comments: comments,
	}
}

func newCommentsLoadFailedMsg(item domain.Item, err error) commentsMsg {
	return commentsMsg{
		baseMsg: baseMsg{state: stateLoadFailed, err: err},
		item:    item,
	}
}

type state int

type baseMsg struct {
	state state
	err   error
}

func (b baseMsg) isLoading() bool {
	return b.state == stateLoading
}

func (b baseMsg) isSuccess() bool {
	return b.state == stateLoadSuccess
}
