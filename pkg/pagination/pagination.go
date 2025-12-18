package pagination

type PaginationParams struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type Pagination struct {
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalPages  int  `json:"totalPages"`
	HasNextPage bool `json:"hasNextPage"`
	HasPrevPage bool `json:"hasPrevPage"`
}

func NewPagination(params *PaginationParams) Pagination {
	return Pagination{
		Total:       params.Total,
		Page:        params.Page,
		Limit:       params.Limit,
		TotalPages:  params.Total / params.Limit,
		HasNextPage: params.Page < params.Total/params.Limit,
		HasPrevPage: params.Page > 1,
	}
}
