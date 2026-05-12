package domain

type Item struct {
	ID          int64
	Type        string
	By          string
	Time        int64
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
