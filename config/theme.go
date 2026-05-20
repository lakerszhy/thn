package config

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

//nolint:gochecknoglobals // this is a constant theme.
var HackerNewsTheme = Theme{
	TitleBar: TitleBarTheme{
		Border: BorderTheme{
			Style:      lipgloss.RoundedBorder(),
			Color:      lipgloss.Color("#6E6E6E"),
			FocusColor: lipgloss.Color("#FF6600"),
		},
		CategoryColor:         lipgloss.Color("#B4B4B4"),
		CategorySelectedColor: lipgloss.Color("#FF6600"),
		DivideColor:           lipgloss.Color("#8C8C8C"),
	},

	Item: ItemTheme{
		TitleColor:          lipgloss.Color("#E6E6E6"),
		TitleSelectedColor:  lipgloss.Color("#FF6600"),
		DomainColor:         lipgloss.Color("#8C8C8C"),
		DomainSelectedColor: lipgloss.Color("#A54301"),
		DescColor:           lipgloss.Color("#8C8C8C"),
		DescSelectedColor:   lipgloss.Color("#A54301"),
	},

	Comment: CommentTheme{
		Border: BorderTheme{
			Style:      lipgloss.RoundedBorder(),
			Color:      lipgloss.Color("#6E6E6E"),
			FocusColor: lipgloss.Color("#FF6600"),
		},
		DescColor:         lipgloss.Color("#8C8C8C"),
		DescSelectedColor: lipgloss.Color("#FF6600"),
		ContentColor:      lipgloss.Color("#E6E6E6"),
	},
}

type Theme struct {
	TitleBar TitleBarTheme
	Item     ItemTheme
	Comment  CommentTheme
}

type TitleBarTheme struct {
	Border                BorderTheme
	CategoryColor         color.Color
	CategorySelectedColor color.Color
	DivideColor           color.Color
}

type ItemTheme struct {
	TitleColor          color.Color
	TitleSelectedColor  color.Color
	DomainColor         color.Color
	DomainSelectedColor color.Color
	DescColor           color.Color
	DescSelectedColor   color.Color
}

type CommentTheme struct {
	Border            BorderTheme
	DescColor         color.Color
	DescSelectedColor color.Color
	ContentColor      color.Color
}

type BorderTheme struct {
	Style      lipgloss.Border
	Color      color.Color
	FocusColor color.Color
}
