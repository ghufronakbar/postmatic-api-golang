package business_knowledge

type UpsertBusinessKnowledgeInput struct {
	PrimaryLogoUrl     string  `json:"primaryLogoUrl" validate:"required,url"`
	Name               string  `json:"name" validate:"required"`
	Category           string  `json:"category" validate:"required"`
	Description        string  `json:"description" validate:"required"`
	UniqueSellingPoint string  `json:"uniqueSellingPoint" validate:"required"`
	WebsiteUrl         *string `json:"websiteUrl" validate:"omitempty,url"`
	VisionMission      string  `json:"visionMission" validate:"required"`
	Location           string  `json:"location" validate:"required"`
	ColorTone          string  `json:"colorTone" validate:"required,min=6,max=6"`
}
