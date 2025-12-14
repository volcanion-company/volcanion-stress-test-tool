package repository

// Pagination holds pagination parameters
type Pagination struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// PaginatedResult holds paginated results
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// DefaultPagination returns default pagination settings
func DefaultPagination() Pagination {
	return Pagination{
		Page:     1,
		PageSize: 20,
	}
}

// NewPagination creates pagination with validation
func NewPagination(page, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size
	}
	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset returns the SQL offset for pagination
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the SQL limit for pagination
func (p Pagination) Limit() int {
	return p.PageSize
}

// NewPaginatedResult creates a paginated result
func NewPaginatedResult[T any](items []T, total int64, pagination Pagination) PaginatedResult[T] {
	totalPages := int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginatedResult[T]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}
}

// ListOptions provides filtering and pagination options
type ListOptions struct {
	Pagination Pagination
	SortBy     string
	SortOrder  string // "asc" or "desc"
	Filters    map[string]interface{}
}

// DefaultListOptions returns default list options
func DefaultListOptions() ListOptions {
	return ListOptions{
		Pagination: DefaultPagination(),
		SortBy:     "created_at",
		SortOrder:  "desc",
		Filters:    make(map[string]interface{}),
	}
}

// WithPage sets the page number
func (o ListOptions) WithPage(page int) ListOptions {
	o.Pagination.Page = page
	if o.Pagination.Page < 1 {
		o.Pagination.Page = 1
	}
	return o
}

// WithPageSize sets the page size
func (o ListOptions) WithPageSize(size int) ListOptions {
	o.Pagination.PageSize = size
	if o.Pagination.PageSize < 1 {
		o.Pagination.PageSize = 20
	}
	if o.Pagination.PageSize > 100 {
		o.Pagination.PageSize = 100
	}
	return o
}

// WithSort sets sorting options
func (o ListOptions) WithSort(sortBy, sortOrder string) ListOptions {
	o.SortBy = sortBy
	o.SortOrder = sortOrder
	if o.SortOrder != "asc" && o.SortOrder != "desc" {
		o.SortOrder = "desc"
	}
	return o
}

// WithFilter adds a filter
func (o ListOptions) WithFilter(key string, value interface{}) ListOptions {
	if o.Filters == nil {
		o.Filters = make(map[string]interface{})
	}
	o.Filters[key] = value
	return o
}
