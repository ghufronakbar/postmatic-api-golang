package business_role

import "time"

type BusinessRoleResponse struct {
	BusinessRootId  string    `json:"businessRootId"`
	AudiencePersona string    `json:"audiencePersona"`
	CallToAction    string    `json:"callToAction"`
	Goals           string    `json:"goals"`
	Hashtags        []string  `json:"hashtags"`
	TargetAudience  string    `json:"targetAudience"`
	Tone            string    `json:"tone"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
