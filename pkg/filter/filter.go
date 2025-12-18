// pkg/filter/filter.go
package filter

type ReqFilter struct {
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	Sort     string `json:"sort"`   // asc|desc
	SortBy   string `json:"sortBy"` // name|createdAt|updatedAt (input user)
	Category string `json:"category"`
}

// Offset adalah nilai turunan (hasil kalkulasi), bukan input user
func (f ReqFilter) Offset() int {
	if f.Page < 1 {
		return 0
	}
	if f.Limit < 1 {
		return 0
	}
	return (f.Page - 1) * f.Limit
}

// SortByDB mengubah input user ke nama kolom yang dipakai SQL
func (f ReqFilter) SortByDB() string {
	switch f.SortBy {
	case "name":
		return "name"
	case "createdAt":
		return "created_at"
	case "updatedAt":
		return "updated_at"
	default:
		return "created_at"
	}
}
