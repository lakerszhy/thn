package comments

import "github.com/lakerszhy/thn/domain"

const (
	stateLoading = iota
	stateLoadSuccess
	stateLoadFailed
	// TODO: can delete these states?
	stateLoadingChildren
	stateLoadChildrenSuccess
	stateLoadChildrenFailed
)

type state int

type itemMsg struct {
	itemID int64
	item   domain.Item
	state  state
	err    error
}

func newItemLoadingMsg(itemID int64) itemMsg {
	return itemMsg{
		state:  stateLoading,
		itemID: itemID,
	}
}

func newItemLoadSuccessMsg(itemID int64, item domain.Item) itemMsg {
	return itemMsg{
		state:  stateLoadSuccess,
		itemID: itemID,
		item:   item,
	}
}

func newItemLoadFailedMsg(itemID int64, err error) itemMsg {
	return itemMsg{
		state:  stateLoadFailed,
		err:    err,
		itemID: itemID,
	}
}

type commentsMsg struct {
	item     domain.Item
	parentID int64
	comments []domain.Comment
	state    state
	err      error
}

func newCommentsLoadingMsg(item domain.Item) commentsMsg {
	return commentsMsg{
		item:     item,
		state:    stateLoading,
		parentID: item.ID,
	}
}

func newCommentsLoadSuccessMsg(item domain.Item, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		item:     item,
		parentID: item.ID,
		state:    stateLoadSuccess,
		comments: comments,
	}
}

func newCommentsLoadFailedMsg(item domain.Item, err error) commentsMsg {
	return commentsMsg{
		item:     item,
		parentID: item.ID,
		state:    stateLoadFailed,
		err:      err,
	}
}

func newCommentsLoadingChildrenMsg(item domain.Item, parentID int64) commentsMsg {
	return commentsMsg{
		item:     item,
		parentID: parentID,
		state:    stateLoadingChildren,
	}
}

func newCommentsLoadChildrenSuccessMsg(item domain.Item, parentID int64, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		item:     item,
		state:    stateLoadChildrenSuccess,
		parentID: parentID,
		comments: comments,
	}
}

func newCommentsLoadChildrenFailedMsg(item domain.Item, parentID int64, err error) commentsMsg {
	return commentsMsg{
		item:     item,
		state:    stateLoadChildrenFailed,
		err:      err,
		parentID: parentID,
	}
}
