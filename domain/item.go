package domain

import (
	"time"

	"github.com/dustin/go-humanize"
)

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
