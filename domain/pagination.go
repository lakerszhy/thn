package domain

const defaultPageSize = 50

// TODO: unexported?
type Pagination struct {
	page       int // start at 0
	pageSize   int
	totalCount int
}

func NewPagination() Pagination {
	return Pagination{page: 0, pageSize: defaultPageSize}
}

func (p Pagination) SetTotalCount(v int) Pagination {
	p.totalCount = v
	return p
}

func (p Pagination) Next() Pagination {
	p.page++
	return p
}

func (p Pagination) HasMore() bool {
	return (p.page+1)*p.pageSize < p.totalCount
}

func (p Pagination) Range() (int, int) {
	start := p.page * p.pageSize

	if start >= p.totalCount || start < 0 {
		return p.totalCount, p.totalCount
	}

	end := min(start+p.pageSize, p.totalCount)

	return start, end
}
