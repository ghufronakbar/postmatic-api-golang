package business_knowledge

type BusinessKnowledgeResponse struct {
	PrimaryLogoUrl     string  `json:"primaryLogoUrl"`
	Name               string  `json:"name"`
	Category           string  `json:"category"`
	Description        string  `json:"description"`
	UniqueSellingPoint string  `json:"uniqueSellingPoint"`
	WebsiteUrl         *string `json:"websiteUrl"`
	VisionMission      string  `json:"visionMission"`
	Location           string  `json:"location"`
	RootBusinessId     string  `json:"rootBusinessId"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
	ColorTone          string  `json:"colorTone"`
}
