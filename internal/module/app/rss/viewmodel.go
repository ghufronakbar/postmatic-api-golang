package rss

import "time"

type RSSResponse struct {
	ID                  string    `json:"id"`
	Title               string    `json:"title"`
	URL                 string    `json:"url"`
	Publisher           string    `json:"publisher"`
	MasterRSSCategoryID string    `json:"masterRssCategoryId"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type RSSCategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
