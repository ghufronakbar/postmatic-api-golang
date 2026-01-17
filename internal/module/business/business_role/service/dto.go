package business_role_service

type UpsertBusinessRoleInput struct {
	AudiencePersona string   `json:"audiencePersona" validate:"required"`
	CallToAction    string   `json:"callToAction" validate:"required"`
	Goals           string   `json:"goals" validate:"required"`
	Hashtags        []string `json:"hashtags" validate:"required"`
	TargetAudience  string   `json:"targetAudience" validate:"required"`
	Tone            string   `json:"tone" validate:"required"`
	BusinessRootID  int64    `json:"businessRootId" validate:"required"`
}
