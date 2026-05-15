package config

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var DefaultTheme = Theme{
	Border: border{
		Style:       lipgloss.RoundedBorder(),
		Color:       lipgloss.Color("#757075"),
		ActiveColor: lipgloss.Color("#baebf6"),
	},

	CategoryColor:       lipgloss.Color("#fcfcfc"),
	CategoryActiveColor: lipgloss.Color("#64d2e8"),

	ItemTitleColor:         lipgloss.Color("#fcfcfc"),
	ItemTitleSelectedColor: lipgloss.Color("#c6e472"),
	ItemDescColor:          lipgloss.Color("#fcfcfc"),
	ItemDescSelectedColor:  lipgloss.Color("#c6e472"),

	CommentDescColor:    lipgloss.Color("#fcfcfc"),
	CommentContentColor: lipgloss.Color("#fcfcfc"),
}

type Theme struct {
	Border border

	CategoryColor       color.Color
	CategoryActiveColor color.Color

	ItemTitleColor         color.Color
	ItemTitleSelectedColor color.Color
	ItemDescColor          color.Color
	ItemDescSelectedColor  color.Color

	CommentDescColor    color.Color
	CommentContentColor color.Color
}

type border struct {
	Style       lipgloss.Border
	Color       color.Color
	ActiveColor color.Color
}
