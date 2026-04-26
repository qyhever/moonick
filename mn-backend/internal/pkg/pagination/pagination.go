package pagination

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type Params struct {
	Page     int
	PageSize int
}

func Normalize(page int, pageSize int) Params {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Params{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}
