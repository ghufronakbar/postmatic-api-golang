package business_knowledge_service

import "time"

type BusinessKnowledgeResponse struct {
	PrimaryLogoUrl     string    `json:"primaryLogoUrl"`
	Name               string    `json:"name"`
	Category           string    `json:"category"`
	Description        string    `json:"description"`
	UniqueSellingPoint string    `json:"uniqueSellingPoint"`
	WebsiteUrl         *string   `json:"websiteUrl"`
	VisionMission      string    `json:"visionMission"`
	Location           string    `json:"location"`
	RootBusinessId     int64     `json:"rootBusinessId"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	ColorTone          string    `json:"colorTone"`
}
