package config

import "charm.land/bubbles/v2/key"

//nolint:gochecknoglobals // default hotkey set
var Hotkeys = Hotkey{
	PrevCategory: key.NewBinding(key.WithKeys("shift+tab")),
	NextCategory: key.NewBinding(key.WithKeys("tab")),

	PrevItem:     key.NewBinding(key.WithKeys("up", "k")),
	NextItem:     key.NewBinding(key.WithKeys("down", "j")),
	PrevPage:     key.NewBinding(key.WithKeys("left", "h")),
	NextPage:     key.NewBinding(key.WithKeys("right", "l")),
	OpenComments: key.NewBinding(key.WithKeys("enter", "space")),

	CloseComments:      key.NewBinding(key.WithKeys("esc")),
	ToggleFullScreen:   key.NewBinding(key.WithKeys("ctrl+enter")),
	ToggleSubComments:  key.NewBinding(key.WithKeys("enter", "space")),
	SelectParent:       key.NewBinding(key.WithKeys("left", "h")),
	PrevComment:        key.NewBinding(key.WithKeys("up", "k")),
	NextComment:        key.NewBinding(key.WithKeys("down", "j")),
	PrevSiblingComment: key.NewBinding(key.WithKeys("p")),
	NextSiblingComment: key.NewBinding(key.WithKeys("n")),

	Refresh: key.NewBinding(key.WithKeys("r")),

	OpenHNInBrowser:      key.NewBinding(key.WithKeys("o")),
	OpenCommentInBrowser: key.NewBinding(key.WithKeys("ctrl+o")),
	OpenDomainInBrowser:  key.NewBinding(key.WithKeys("O")),

	GoToStart: key.NewBinding(key.WithKeys("home", "g")),
	GoToEnd:   key.NewBinding(key.WithKeys("end", "G")),
	Quit:      key.NewBinding(key.WithKeys("ctrl+c", "q")),
}

type Hotkey struct {
	PrevCategory key.Binding
	NextCategory key.Binding

	PrevItem     key.Binding
	NextItem     key.Binding
	PrevPage     key.Binding
	NextPage     key.Binding
	OpenComments key.Binding

	CloseComments      key.Binding
	ToggleFullScreen   key.Binding
	ToggleSubComments  key.Binding
	SelectParent       key.Binding
	PrevComment        key.Binding
	NextComment        key.Binding
	PrevSiblingComment key.Binding
	NextSiblingComment key.Binding

	Refresh key.Binding

	OpenHNInBrowser      key.Binding
	OpenCommentInBrowser key.Binding
	OpenDomainInBrowser  key.Binding

	GoToStart key.Binding
	GoToEnd   key.Binding
	Quit      key.Binding
}
