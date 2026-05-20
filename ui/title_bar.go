package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/lakerszhy/thn/config"
	"github.com/lakerszhy/thn/domain"
)

type titleBar struct {
	categories []domain.Category
	theme      config.TitleBarTheme
	current    domain.Category
	width      int
}

func newTitleBar(categories []domain.Category, theme config.TitleBarTheme) *titleBar {
	return &titleBar{
		categories: categories,
		theme:      theme,
	}
}

func (t *titleBar) View() string {
	catStyle := lipgloss.NewStyle().Padding(0, 1)

	categories := make([]string, len(t.categories))
	for i, c := range t.categories {
		if c == t.current {
			catStyle = catStyle.Foreground(t.theme.CategorySelectedColor).Bold(true)
		} else {
			catStyle = catStyle.Foreground(t.theme.CategoryColor)
		}
		categories[i] = catStyle.Render(string(c))
	}

	divider := lipgloss.NewStyle().Foreground(t.theme.DivideColor).Render("|")

	return lipgloss.NewStyle().Width(t.width).Render(strings.Join(categories, divider))
}

func (t *titleBar) setWidth(width int) {
	t.width = width
}

func (t *titleBar) setCategory(cat domain.Category) {
	t.current = cat
}
