// internal/module/generative_content/image_post/service/dto.go
package image_post_service

import "time"

// CreateImagePostInput adalah input untuk membuat image post
type CreateImagePostInput struct {
	BusinessRootID            int64  `json:"-"`
	Mode                      string `json:"mode" validate:"required,oneof=generate regenerate rss mask"`
	Ratio                     string `json:"ratio" validate:"required"`
	ProductKnowledgeID        int64  `json:"productKnowledgeId" validate:"required,gte=1"`
	AppGenerativeImageModelID int64  `json:"appGenerativeImageModelId" validate:"required,gte=1"`
	NumOfImages               int    `json:"numOfImages" validate:"required,gte=1,lte=4"`

	// Optional
	AdditionalPrompt string `json:"additionalPrompt"`
	DesignStyle      string `json:"designStyle"`
	Category         string `json:"category"`
	ReferenceImage   string `json:"referenceImage"`
	MaskImage        string `json:"maskImage"`
	ImageSize        string `json:"imageSize"`
	CurrentCaption   string `json:"currentCaption"`

	// Advance Generate
	AdvanceGenerate *AdvanceGenerateInput `json:"advanceGenerate"`

	// RSS
	Rss *RssInput `json:"rss"`
}

// AdvanceGenerateInput untuk memilih knowledge mana yang digunakan
type AdvanceGenerateInput struct {
	BusinessKnowledge *BusinessKnowledgeFlags `json:"businessKnowledge"`
	ProductKnowledge  *ProductKnowledgeFlags  `json:"productKnowledge"`
	RoleKnowledge     *RoleKnowledgeFlags     `json:"roleKnowledge"`
}

type BusinessKnowledgeFlags struct {
	Name               bool `json:"name"`
	Category           bool `json:"category"`
	Description        bool `json:"description"`
	Location           bool `json:"location"`
	Logo               bool `json:"logo"`
	UniqueSellingPoint bool `json:"uniqueSellingPoint"`
	Website            bool `json:"website"`
	VisionMission      bool `json:"visionMission"`
	ColorTone          bool `json:"colorTone"`
}

type ProductKnowledgeFlags struct {
	Name        bool `json:"name"`
	Category    bool `json:"category"`
	Description bool `json:"description"`
	Price       bool `json:"price"`
}

type RoleKnowledgeFlags struct {
	Hashtags bool `json:"hashtags"`
}

// RssInput untuk mode RSS
type RssInput struct {
	Title       string     `json:"title" validate:"required"`
	URL         string     `json:"url" validate:"required,url"`
	PublishedAt *time.Time `json:"publishedAt"`
	ImageURL    *string    `json:"imageUrl"`
	Summary     string     `json:"summary"`
	Publisher   string     `json:"publisher"`
}

// GetImagePostsFilter untuk list dengan pagination
type GetImagePostsFilter struct {
	BusinessRootID    int64
	Status            *string
	Mode              *string
	BusinessProductID *int64
	SortBy            string
	SortDir           string
	PageOffset        int
	PageLimit         int
	Page              int
	DateStart         *string
	DateEnd           *string
}
