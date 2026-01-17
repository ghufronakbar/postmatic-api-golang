package generative_image_model_service

type GetGenerativeImageModelsFilter struct {
	Search     string `json:"search"`
	SortBy     string `json:"sortBy"`
	SortDir    string `json:"sortDir"`
	PageOffset int    `json:"pageOffset"`
	PageLimit  int    `json:"pageLimit"`
	Page       int    `json:"page"`
	IsAdmin    bool   `json:"isAdmin"`
}

var SORT_BY = []string{"label", "model", "created_at", "updated_at", "id"}
