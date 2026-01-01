// internal/module/business/business_member/viewmodel.go
package business_member

import (
	"time"

	"github.com/google/uuid"
)

type BusinessMemberResponse struct {
	ID         int64              `json:"id"`
	Role       string             `json:"role"`
	Status     string             `json:"status"`
	Profile    BusinessProfileSub `json:"profile"`
	AnsweredAt *time.Time         `json:"answeredAt"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
	IsYourself bool               `json:"isYourself"`
}

type BusinessProfileSub struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Image *string   `json:"image"`
	ID    uuid.UUID `json:"id"`
}

type BusinessMemberInvitationResponse struct {
	ProfileName    string    `json:"profileName"`
	ProfileEmail   string    `json:"profileEmail"`
	ProfileID      uuid.UUID `json:"profileId"`
	ProfileImage   string    `json:"profileImage"`
	BusinessName   string    `json:"businessName"`
	BusinessRootID int64     `json:"businessRootId"`
	BusinessLogo   string    `json:"businessLogo"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	Valid          bool      `json:"valid"`
	ExpiredAt      time.Time `json:"expiredAt"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type GeneralMemberResponse struct {
	ID             int64     `json:"id"`
	BusinessRootID int64     `json:"businessRootId"`
	ProfileID      uuid.UUID `json:"profileId"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type InviteMemberResponse struct {
	GeneralMemberResponse
	InvitationLink string `json:"invitationLink"`
	RetryAfter     int64  `json:"retryAfter"`
}
