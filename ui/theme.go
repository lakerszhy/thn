package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var DefaultTheme = Theme{
	border: border{
		style:       lipgloss.RoundedBorder(),
		color:       lipgloss.Color("#757075"),
		activeColor: lipgloss.Color("#baebf6"),
	},

	categoryColor:       lipgloss.Color("#fcfcfc"),
	categoryActiveColor: lipgloss.Color("#64d2e8"),

	itemTitleColor:         lipgloss.Color("#fcfcfc"),
	itemTitleSelectedColor: lipgloss.Color("#c6e472"),
	itemDescColor:          lipgloss.Color("#fcfcfc"),
	itemDescSelectedColor:  lipgloss.Color("#c6e472"),

	commentDescColor:    lipgloss.Color("#fcfcfc"),
	commentContentColor: lipgloss.Color("#fcfcfc"),
}

type Theme struct {
	border border

	categoryColor       color.Color
	categoryActiveColor color.Color

	itemTitleColor         color.Color
	itemTitleSelectedColor color.Color
	itemDescColor          color.Color
	itemDescSelectedColor  color.Color

	commentDescColor    color.Color
	commentContentColor color.Color
}

type border struct {
	style       lipgloss.Border
	color       color.Color
	activeColor color.Color
}
