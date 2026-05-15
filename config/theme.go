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

	ItemColor:         lipgloss.Color("#fcfcfc"),
	ItemSelectedColor: lipgloss.Color("#c6e472"),

	CommentDescColor:    lipgloss.Color("#fcfcfc"),
	CommentContentColor: lipgloss.Color("#fcfcfc"),
}

type Theme struct {
	Border border

	CategoryColor       color.Color
	CategoryActiveColor color.Color

	ItemColor         color.Color
	ItemSelectedColor color.Color

	CommentDescColor    color.Color
	CommentContentColor color.Color
}

type border struct {
	Style       lipgloss.Border
	Color       color.Color
	ActiveColor color.Color
}
