// internal/module/business/business_information/viewmodel.go
package business_information

import "time"

type SetupBusinessRootFirstTimeResponse struct {
	// Business Root ID
	ID string `json:"id"`
}

type GetJoinedBusinessesByProfileIDResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
	AnsweredAt   time.Time           `json:"answeredAt"`
	Members      []BusinessMemberSub `json:"members"`
	UserPosition BusinessMemberSub   `json:"userPosition"`
}

type BusinessMemberSub struct {
	Status  string     `json:"status"`
	Role    string     `json:"role"`
	Profile ProfileSub `json:"profile"`
}

type ProfileSub struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ImageUrl *string `json:"imageUrl"`
	Email    string  `json:"email"`
}

type GetBusinessByIdResponse struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	PrimaryLogoUrl *string             `json:"primaryLogoUrl"`
	Category       string              `json:"category"`
	Description    *string             `json:"description"`
	ColorTone      *string             `json:"colorTone"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
	Members        []BusinessMemberSub `json:"members"`
	UserPosition   BusinessMemberSub   `json:"userPosition"`
}
