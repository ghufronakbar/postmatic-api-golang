// internal/module/business/business_information/dto.go
package business_information_service

import "github.com/google/uuid"

type BusinessSetupInput struct {
	BusinessKnowledge BusinessKnowledgeSub `json:"businessKnowledge" validate:"required"`
	ProductKnowledge  ProductKnowledgeSub  `json:"productKnowledge" validate:"required"`
	RoleKnowledge     RoleKnowledgeSub     `json:"roleKnowledge" validate:"required"`
	ProfileID         uuid.UUID            `json:"profileId" validate:"required"`
}

type BusinessKnowledgeSub struct {
	Name               string `json:"name" validate:"required"`
	PrimaryLogoUrl     string `json:"primaryLogoUrl" validate:"required,url"`
	Category           string `json:"category" validate:"required"`
	Description        string `json:"description" validate:"required"`
	UniqueSellingPoint string `json:"uniqueSellingPoint" validate:"required"`
	WebsiteUrl         string `json:"websiteUrl" validate:"omitempty,url"`
	VisionMission      string `json:"visionMission" validate:"required"`
	Location           string `json:"location" validate:"required"`
	ColorTone          string `json:"colorTone" validate:"required,min=6,max=6"`
}

type ProductKnowledgeSub struct {
	Name        string   `json:"name" validate:"required"`
	Category    string   `json:"category" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Price       int64    `json:"price" validate:"required"`
	Currency    string   `json:"currency" validate:"required"`
	ImageUrls   []string `json:"imageUrls" validate:"required"`
}

type RoleKnowledgeSub struct {
	TargetAudience  string   `json:"targetAudience" validate:"required"`
	Tone            string   `json:"tone" validate:"required"`
	AudiencePersona string   `json:"audiencePersona" validate:"required"`
	Hashtags        []string `json:"hashtags" validate:"required"`
	CallToAction    string   `json:"callToAction" validate:"required"`
	Goals           string   `json:"goals" validate:"required"`
}
