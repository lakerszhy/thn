package ui

import "github.com/lakerszhy/thn/domain"

type itemsMsg struct {
	category domain.Category
	items    []domain.Item
}
