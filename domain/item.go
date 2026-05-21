package domain

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

const (
	ItemTypeStory ItemType = iota
	ItemTypeComment
	ItemTypeJob
	ItemTypePoll
	ItemTypePollOpt
)

type ItemType int

type PagedItems struct {
	Pagination

	Items []Item
}

func NewPagedItems(p Pagination, items []Item) PagedItems {
	return PagedItems{Pagination: p, Items: items}
}

type Item struct {
	Base

	Poll        int64
	KIDs        []int64
	URL         string
	Score       int64
	Title       string
	Parts       []int64
	Descendants int64
}

func (i Item) HasComments() bool {
	return len(i.KIDs) > 0
}

func (i Item) Description() string {
	v := fmt.Sprintf("by %s %s", i.By, i.TimeAgo())

	if i.Type != ItemTypeJob {
		v = fmt.Sprintf("%d points %s", i.Score, v)
	}

	if i.Descendants == 1 {
		v = fmt.Sprintf("%s | 1 comment", v)
	} else if i.Descendants > 1 {
		v = fmt.Sprintf("%s | %d comments", v, i.Descendants)
	}

	return v
}

func (i Item) Domain() string {
	if i.URL == "" {
		return ""
	}

	u, err := url.Parse(i.URL)
	if err != nil {
		slog.Error("fail to parse url", "id", i.ID, "url", i.URL, "error", err)
		return ""
	}

	host := strings.TrimPrefix(u.Hostname(), "www.")

	if host == "github.com" || host == "twitter.com" || host == "x.com" {
		paths := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(paths) > 1 {
			r, _ := url.JoinPath(host, paths[0])
			return r
		}
	}

	return host
}

type Comment struct {
	Base

	Parent      int64
	Descendants int64
	KIDs        []int64
}

func (c Comment) URLInHN(itemID int64) string {
	return fmt.Sprintf("%s#%d", hnURL(itemID), c.ID)
}

type Base struct {
	ID      int64
	Time    int64
	By      string
	Type    ItemType
	Text    string
	Deleted bool
	Dead    bool
}

func (b Base) TimeAgo() string {
	return humanize.Time(time.Unix(b.Time, 0))
}

func (b Base) URLInHN() string {
	return hnURL(b.ID)
}

func hnURL(id int64) string {
	return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", id)
}
