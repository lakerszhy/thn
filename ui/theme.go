package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var DefaultTheme = Theme{
	border: lipgloss.RoundedBorder(),

	categoryBackgroundColor: lipgloss.Color("#ff6600"),
	categoryColor:           lipgloss.Color("#000000"),
	categoryActiveColor:     lipgloss.Color("#ffffff"),
}

type Theme struct {
	border lipgloss.Border

	categoryBackgroundColor color.Color
	categoryColor           color.Color
	categoryActiveColor     color.Color
}
