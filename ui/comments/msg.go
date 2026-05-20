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

// If commentID is zero, it is the comments of the item whose id is itemID.
// If commentID is not zero, it is sub comments of the comment whose id is *commentID.
type commentsMsg struct {
	itemID    int64
	commentID int64
	comments  []domain.Comment
	state     state
	err       error
}

func (m commentsMsg) isRoot() bool {
	return m.commentID == 0
}

func newCommentsLoadingMsg(itemID int64) commentsMsg {
	return commentsMsg{
		itemID: itemID,
		state:  stateLoading,
	}
}

func newCommentsLoadSuccessMsg(itemID int64, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		itemID:   itemID,
		state:    stateLoadSuccess,
		comments: comments,
	}
}

func newCommentsLoadFailedMsg(itemID int64, err error) commentsMsg {
	return commentsMsg{
		itemID: itemID,
		state:  stateLoadFailed,
		err:    err,
	}
}

func newSubCommentsLoadingMsg(itemID int64, commentID int64) commentsMsg {
	return commentsMsg{
		itemID:    itemID,
		commentID: commentID,
		state:     stateLoading,
	}
}

func newSubCommentsLoadSuccessMsg(itemID int64, commentID int64, comments []domain.Comment) commentsMsg {
	return commentsMsg{
		itemID:    itemID,
		commentID: commentID,
		state:     stateLoadSuccess,
		comments:  comments,
	}
}

func newSubCommentsLoadFailedMsg(itemID int64, commentID int64, err error) commentsMsg {
	return commentsMsg{
		itemID:    itemID,
		commentID: commentID,
		state:     stateLoadFailed,
		err:       err,
	}
}
