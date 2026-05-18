package config

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var HackerNewsTheme = Theme{
	TitleBar: TitleBarTheme{
		Border: BorderTheme{
			Style:      lipgloss.RoundedBorder(),  // 圆角边框更有现代感
			Color:      lipgloss.Color("#6E6E6E"), // 未激活时暗灰边框
			FocusColor: lipgloss.Color("#FF6600"), // 激活/聚焦时 HN 橙色边框
		},
		CategoryColor:         lipgloss.Color("#B4B4B4"), // 浅灰导航
		CategorySelectedColor: lipgloss.Color("#FF6600"), // 选中分类变橙
		DivideColor:           lipgloss.Color("#8C8C8C"),
	},

	Item: ItemTheme{
		TitleColor:          lipgloss.Color("#E6E6E6"), // 亮白/浅灰标题，护眼
		TitleSelectedColor:  lipgloss.Color("#FF6600"), // 选中行变橙色
		DomainColor:         lipgloss.Color("#8C8C8C"), // 暗灰域名
		DomainSelectedColor: lipgloss.Color("#a54301"), // 选中行域名变成过渡橙灰
		DescColor:           lipgloss.Color("#8C8C8C"), // 灰色次要信息
		DescSelectedColor:   lipgloss.Color("#a54301"), // 选中行次要信息变亮
	},

	Comment: CommentTheme{
		DescColor:         lipgloss.Color("#8C8C8C"), // 评论区作者与时间
		DescSelectedColor: lipgloss.Color("#FF6600"),
		ContentColor:      lipgloss.Color("#E6E6E6"), // 评论正文
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
	DescColor         color.Color
	DescSelectedColor color.Color
	ContentColor      color.Color
}

type BorderTheme struct {
	Style      lipgloss.Border
	Color      color.Color
	FocusColor color.Color
}
