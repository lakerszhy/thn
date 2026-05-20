package domain

import (
	"net/url"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type PagedItems struct {
	Pagination

	Items []Item
}

func NewPagedItems(p Pagination, items []Item) PagedItems {
	return PagedItems{Pagination: p, Items: items}
}

type Item struct {
	Base

	Type        string
	Text        string
	Parent      int64
	Poll        int64
	KIDs        []int64
	URL         string
	Score       int64
	Title       string
	Parts       []int64
	Descendants int64
}

func (i Item) Domain() string {
	if i.URL == "" {
		return ""
	}

	u, err := url.Parse(i.URL)
	if err != nil {
		// TODO: should log
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

	Text        string
	Parent      int64
	Descendants int64
	KIDs        []int64
}

type Base struct {
	ID      int64
	Time    int64
	By      string
	Deleted bool
	Dead    bool
}

func (b Base) TimeAgo() string {
	return humanize.Time(time.Unix(b.Time, 0))
}
