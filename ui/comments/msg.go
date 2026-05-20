package comments

import "github.com/lakerszhy/thn/domain"

const (
	stateLoading = iota
	stateLoadSuccess
	stateLoadFailed
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

// If parentID == itemID, it is the comments of the item whose id is itemID.
// If parentID != itemID, it is sub comments of the comment whose id is parentID.
type commentsMsg struct {
	itemID   int64
	parentID int64
	comments []domain.Comment
	state    state
	err      error
}

func (m commentsMsg) isRoot() bool {
	return m.parentID == m.itemID
}

func newCommentsLoadingMsg(itemID int64) commentsMsg {
	return commentsMsg{
		itemID:   itemID,
		parentID: itemID,
		state:    stateLoading,
	}
}

func newCommentsLoadSuccessMsg(itemID int64, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		itemID:   itemID,
		parentID: itemID,
		state:    stateLoadSuccess,
		comments: comments,
	}
}

func newCommentsLoadFailedMsg(itemID int64, err error) commentsMsg {
	return commentsMsg{
		itemID:   itemID,
		parentID: itemID,
		state:    stateLoadFailed,
		err:      err,
	}
}

func newSubCommentsLoadingMsg(itemID int64, parentID int64) commentsMsg {
	return commentsMsg{
		itemID:   itemID,
		parentID: parentID,
		state:    stateLoading,
	}
}

func newSubCommentsLoadSuccessMsg(itemID int64, parentID int64, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		itemID:   itemID,
		parentID: parentID,
		state:    stateLoadSuccess,
		comments: comments,
	}
}

func newSubCommentsLoadFailedMsg(itemID int64, parentID int64, err error) commentsMsg {
	return commentsMsg{
		itemID:   itemID,
		parentID: parentID,
		state:    stateLoadFailed,
		err:      err,
	}
}
