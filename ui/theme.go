package ui

import "charm.land/lipgloss/v2"

var DefaultTheme = Theme{
	border: lipgloss.RoundedBorder(),
}

type Theme struct {
	border lipgloss.Border
}
