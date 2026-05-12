package hn

import "github.com/lakerszhy/thn/domain"

type category string

func categoryFromDomain(cat domain.Category) category {
	switch cat {
	case domain.CategoryTop:
		return "topstories"
	case domain.CategoryNew:
		return "newstories"
	case domain.CategoryBest:
		return "beststories"
	case domain.CategoryAsk:
		return "askstories"
	case domain.CategoryShow:
		return "showstories"
	case domain.CategoryJob:
		return "jobstories"
	default:
		return "topstories"
	}
}

type item struct {
	ID          int64   `json:"id"`
	Deleted     bool    `json:"deleted"`
	Type        string  `json:"type"`
	By          string  `json:"by"`
	Time        int64   `json:"time"`
	Text        string  `json:"text"`
	Dead        bool    `json:"dead"`
	Parent      int64   `json:"parent"`
	Poll        int64   `json:"poll"`
	KIDs        []int64 `json:"kids"`
	URL         string  `json:"url"`
	Score       int64   `json:"score"`
	Title       string  `json:"title"`
	Parts       []int64 `json:"parts"`
	Descendants int64   `json:"descendants"`
}

func (i item) ToDomain() domain.Item {
	return domain.Item{
		ID:          i.ID,
		Type:        i.Type,
		By:          i.By,
		Time:        i.Time,
		Text:        i.Text,
		Parent:      i.Parent,
		Poll:        i.Poll,
		KIDs:        i.KIDs,
		URL:         i.URL,
		Score:       i.Score,
		Title:       i.Title,
		Parts:       i.Parts,
		Descendants: i.Descendants,
	}
}
