package ui

import "github.com/lakerszhy/thn/domain"

const (
	stateLoading = iota
	stateLoadingMore
	stateLoadSuccess
	stateLoadMoreSuccess
	stateLoadFailed
	stateLoadMoreFailed
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

func newItemsLoadingMoreMsg(cat domain.Category) itemsMsg {
	return itemsMsg{
		baseMsg:  baseMsg{state: stateLoadingMore},
		category: cat,
	}
}

func newItemsLoadMoreSuccessMsg(cat domain.Category, items []domain.Item) itemsMsg {
	return itemsMsg{
		baseMsg:  baseMsg{state: stateLoadMoreSuccess},
		category: cat,
		items:    items,
	}
}

func newItemsLoadMoreFailedMsg(cat domain.Category, err error) itemsMsg {
	return itemsMsg{
		baseMsg:  baseMsg{state: stateLoadMoreFailed, err: err},
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

type itemMsg struct {
	itemID int64
	item   domain.Item
	baseMsg
}

func newItemLoadingMsg(itemID int64) itemMsg {
	return itemMsg{
		baseMsg: baseMsg{state: stateLoading},
		itemID:  itemID,
	}
}

func newItemLoadSuccessMsg(itemID int64, item domain.Item) itemMsg {
	return itemMsg{
		baseMsg: baseMsg{state: stateLoadSuccess},
		itemID:  itemID,
		item:    item,
	}
}

func newItemLoadFailedMsg(itemID int64, err error) itemMsg {
	return itemMsg{
		baseMsg: baseMsg{state: stateLoadFailed, err: err},
		itemID:  itemID,
	}
}

type commentsMsg struct {
	item     domain.Item
	parentID int64
	comments []domain.Comment
	baseMsg
}

func newCommentsLoadingMsg(item domain.Item) commentsMsg {
	return commentsMsg{
		baseMsg:  baseMsg{state: stateLoading},
		item:     item,
		parentID: item.ID,
	}
}

func newCommentsLoadSuccessMsg(item domain.Item, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		baseMsg:  baseMsg{state: stateLoadSuccess},
		item:     item,
		parentID: item.ID,
		comments: comments,
	}
}

func newCommentsLoadFailedMsg(item domain.Item, err error) commentsMsg {
	return commentsMsg{
		baseMsg:  baseMsg{state: stateLoadFailed, err: err},
		item:     item,
		parentID: item.ID,
	}
}

func newCommentChildrenLoadingMsg(item domain.Item, parentID int64) commentsMsg {
	return commentsMsg{
		baseMsg:  baseMsg{state: stateLoadingMore},
		item:     item,
		parentID: parentID,
	}
}

func newCommentChildrenLoadSuccessMsg(item domain.Item, parentID int64, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		baseMsg:  baseMsg{state: stateLoadMoreSuccess},
		item:     item,
		parentID: parentID,
		comments: comments,
	}
}

func newCommentChildrenLoadFailedMsg(item domain.Item, parentID int64, err error) commentsMsg {
	return commentsMsg{
		baseMsg:  baseMsg{state: stateLoadMoreFailed, err: err},
		item:     item,
		parentID: parentID,
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

func (b baseMsg) isLoadingMore() bool {
	return b.state == stateLoadingMore
}
