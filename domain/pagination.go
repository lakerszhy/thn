package domain

const defaultPageSize = 50

type Pagination struct {
	page     int // start at 0
	pageSize int
}

func NewPagination() Pagination {
	return Pagination{page: 0, pageSize: defaultPageSize}
}

func (p Pagination) Next() Pagination {
	p.page++
	return p
}

func (p Pagination) Range(total int) (int, int) {
	start := p.page * p.pageSize

	if start >= total || start < 0 {
		return total, total
	}

	end := min(start+p.pageSize, total)

	return start, end
}
