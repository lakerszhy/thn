package ui

import "github.com/lakerszhy/thn/domain"

type itemsMsg struct {
	category domain.Category
	items    []domain.Item
}

type itemSelectedMsg domain.Item

type commentsMsg struct {
	item     domain.Item
	comments []domain.Item
}
