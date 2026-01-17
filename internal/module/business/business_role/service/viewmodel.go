package business_role_service

import "time"

type BusinessRoleResponse struct {
	BusinessRootId  int64     `json:"businessRootId"`
	AudiencePersona string    `json:"audiencePersona"`
	CallToAction    string    `json:"callToAction"`
	Goals           string    `json:"goals"`
	Hashtags        []string  `json:"hashtags"`
	TargetAudience  string    `json:"targetAudience"`
	Tone            string    `json:"tone"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
