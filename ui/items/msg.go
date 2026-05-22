package items

import "github.com/lakerszhy/thn/domain"

const (
	stateLoading = iota
	stateLoadSuccess
	stateLoadFailed
	stateLoadingMore
	stateLoadMoreSuccess
	stateLoadMoreFailed
)

type state int

type ItemSelectedMsg struct {
	Item       domain.Item
	Fullscreen bool
}

func NewItemSelectedMsg(item domain.Item, fullscreen bool) ItemSelectedMsg {
	return ItemSelectedMsg{
		Item:       item,
		Fullscreen: fullscreen,
	}
}

type itemsMsg struct {
	category   domain.Category
	pagedItems domain.PagedItems
	state      state
	err        error
}

func newLoadingMsg(cat domain.Category) itemsMsg {
	return itemsMsg{
		state:    stateLoading,
		category: cat,
	}
}

func newLoadSuccessMsg(cat domain.Category, pagedItems domain.PagedItems) itemsMsg {
	return itemsMsg{
		state:      stateLoadSuccess,
		category:   cat,
		pagedItems: pagedItems,
	}
}

func newLoadFailedMsg(cat domain.Category, err error) itemsMsg {
	return itemsMsg{
		state:    stateLoadFailed,
		err:      err,
		category: cat,
	}
}

func newLoadingMoreMsg(cat domain.Category) itemsMsg {
	return itemsMsg{
		state:    stateLoadingMore,
		category: cat,
	}
}

func newLoadMoreSuccessMsg(cat domain.Category, pagedItems domain.PagedItems) itemsMsg {
	return itemsMsg{
		state:      stateLoadMoreSuccess,
		category:   cat,
		pagedItems: pagedItems,
	}
}

func newLoadMoreFailedMsg(cat domain.Category, err error) itemsMsg {
	return itemsMsg{
		state:    stateLoadMoreFailed,
		err:      err,
		category: cat,
	}
}
