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

	Item: ItemTheme{
		TitleColor:          lipgloss.Color("#fcfcfc"),
		TitleSelectedColor:  lipgloss.Color("#c6e472"),
		DomainColor:         lipgloss.Color("#828282"),
		DomainSelectedColor: lipgloss.Color("#8fa060"),
		DescColor:           lipgloss.Color("#828282"),
		DescSelectedColor:   lipgloss.Color("#8fa060"),
	},

	CategoryColor:       lipgloss.Color("#fcfcfc"),
	CategoryActiveColor: lipgloss.Color("#64d2e8"),

	CommentDescColor:    lipgloss.Color("#fcfcfc"),
	CommentContentColor: lipgloss.Color("#fcfcfc"),
}

type Theme struct {
	Border border

	Item ItemTheme

	CategoryColor       color.Color
	CategoryActiveColor color.Color

	CommentDescColor    color.Color
	CommentContentColor color.Color
}

type ItemTheme struct {
	TitleColor          color.Color
	TitleSelectedColor  color.Color
	DomainColor         color.Color
	DomainSelectedColor color.Color
	DescColor           color.Color
	DescSelectedColor   color.Color
}

type border struct {
	Style       lipgloss.Border
	Color       color.Color
	ActiveColor color.Color
}
