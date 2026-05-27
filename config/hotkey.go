package config

import (
	_ "embed"

	"github.com/pelletier/go-toml/v2"
)

//go:embed templates/hotkeys.toml
var hotkeyTemplate []byte

type Hotkey struct {
	PrevCategory []string `toml:"prev_category"`
	NextCategory []string `toml:"next_category"`

	PrevItem               []string `toml:"prev_item"`
	NextItem               []string `toml:"next_item"`
	PrevPage               []string `toml:"prev_page"`
	NextPage               []string `toml:"next_page"`
	OpenComments           []string `toml:"open_comments"`
	OpenCommentsFullscreen []string `toml:"open_comments_fullscreen"`

	CloseComments      []string `toml:"close_comments"`
	ToggleFullscreen   []string `toml:"toggle_fullscreen"`
	ToggleSubComments  []string `toml:"toggle_sub_comments"`
	SelectParent       []string `toml:"select_parent"`
	PrevComment        []string `toml:"prev_comment"`
	NextComment        []string `toml:"next_comment"`
	PrevSiblingComment []string `toml:"prev_sibling_comment"`
	NextSiblingComment []string `toml:"next_sibling_comment"`

	Refresh []string `toml:"refresh"`

	OpenHNInBrowser      []string `toml:"open_hn_in_browser"`
	OpenCommentInBrowser []string `toml:"open_comment_in_browser"`
	OpenDomainInBrowser  []string `toml:"open_domain_in_browser"`

	GoToStart []string `toml:"go_to_start"`
	GoToEnd   []string `toml:"go_to_end"`
	Quit      []string `toml:"quit"`
}

func LoadHotkeys() (Hotkey, error) {
	var hotkey Hotkey

	err := toml.Unmarshal(hotkeyTemplate, &hotkey)
	if err != nil {
		return hotkey, err
	}

	return hotkey, nil
}
