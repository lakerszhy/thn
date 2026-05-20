package domain

//nolint:gochecknoglobals // this is a constant set of categories
var (
	CategoryTop  Category = "Top"
	CategoryNew  Category = "New"
	CategoryBest Category = "Best"
	CategoryAsk  Category = "Ask"
	CategoryShow Category = "Show"
	CategoryJob  Category = "Job"
)

type Category string
